package avclient

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	// https://github.com/egirna/icap-client
	ic "github.com/integrationworks/icap-client"

	pdk "github.com/Kong/go-pdk"
)

// IKong is an interface to the kong functionality that the plugin uses. The interface greatly eases unit testing.
type IKong interface {
	Debug(args ...interface{}) error
	Info(args ...interface{}) error
	Warn(args ...interface{}) error
	Err(args ...interface{}) error

	ResponseSetHeader(k string, v string) error
	ResponseExit(status int, body string, headers map[string][]string)
}

type DefaultKong struct {
	kong *pdk.PDK
}

func (k DefaultKong) Debug(args ...interface{}) error {
	return k.kong.Log.Debug(args...)
}

func (k DefaultKong) Info(args ...interface{}) error {
	return k.kong.Log.Info(args...)
}

func (k DefaultKong) Warn(args ...interface{}) error {
	return k.kong.Log.Warn(args...)
}

func (k DefaultKong) Err(args ...interface{}) error {
	return k.kong.Log.Err(args...)
}

func (k DefaultKong) ResponseSetHeader(key string, v string) error {
	return k.kong.Response.SetHeader(key, v)
}

func (k DefaultKong) ResponseExit(status int, body string, headers map[string][]string) {
	k.kong.Response.Exit(status, body, headers)
}

/* Access method */
func DoAccess(kong *pdk.PDK, scannerURL string) {
	mykong := DefaultKong{kong}
	mykong.Info("Start Processing Access(kong *pdk.PDK)")

	// Get ContentType this way because it may not be set in the request.
	requestHeaders, err := kong.Request.GetHeaders(100) // TODO check err
	contentType := getContentType(mykong, requestHeaders)

	mykong.Debug("Request contentType:" + contentType)

	// Get the Request body
	body, err := kong.Request.GetRawBody()

	if err != nil {
		mykong.ResponseSetHeader("X-Kong-Virus-Scanned", "false")
		// Can't find a better way to determine that the error is because the file is "too large"
		if err.Error() == "request body did not fit into client body buffer, consider raising 'client_body_buffer_size'" {
			mykong.Warn("Failed to Process Request because it was too large : ",
				"request body did not fit into client body buffer, consider raising 'client_body_buffer_size'")
			mykong.ResponseExit(413, "", nil)
			return
		}

		mykong.Err("Error from GetRawBody:", err.Error())
		mykong.ResponseExit(500, "", nil)
		return

	}

	bodyAsByteArray := []byte(body)

	// mykong.Debug("bodyAsByteArray as bytes:" + fmt.Sprintf("%v", bodyAsByteArray))
	mykong.Debug("bodyAsByteArray as golang:" + fmt.Sprintf("%q", bodyAsByteArray))

	bodyIsOK, err := checkBody(mykong, scannerURL, contentType, body)
	if err != nil {
		mykong.Err("Error from checkBody:", err.Error())
		mykong.Info("Finish processing Access(kong *pdk.PDK). Error from checkBody")
		mykong.ResponseSetHeader("X-Kong-Virus-Scanned", "false")
		mykong.ResponseExit(500, "", nil)
		return
	} else if !bodyIsOK {
		mykong.Warn("Detected a virus")
		mykong.Info("Finish processing Access(kong *pdk.PDK). Virus detected")
		mykong.ResponseSetHeader("X-Kong-Virus-Scanned", "true")
		mykong.ResponseExit(400, "", nil)
		return
	} else {
		mykong.ResponseSetHeader("X-Kong-Virus-Scanned", "true")
		mykong.Info("Finish processing Access(kong *pdk.PDK). No virus detected")
		// Do nothing (just forward the request)
	}

}

func getContentType(mykong IKong, requestHeaders map[string][]string) string {
	contentTypes, ok := requestHeaders["content-type"]
	if !ok {
		mykong.Warn("content-type not set so assuming application/octet-stream")
		return "application/octet-stream"
	}

	return contentTypes[0]
}

func checkBody(mykong IKong, scannerURL string, requestContentType, body string) (bool, error) {
	// TODO Consider using a single return of error

	mykong.Info(fmt.Sprintf("Start checkBody"))
	mediaType, params, err := mime.ParseMediaType(requestContentType)
	if err != nil {
		mykong.Err("Error from mime.ParseMediaType(requestContentType):", err.Error())
		return false, fmt.Errorf("Failed to determine mediaType:%v", requestContentType)
	}

	mykong.Debug(fmt.Sprintf("checkBody():mediaType:%v", mediaType))
	mykong.Debug(fmt.Sprintf("checkBody():params:%v", params))

	// Check Content-Type of whole body
	if strings.HasPrefix(mediaType, "multipart/") {
		// Convert `body` to "files" (bodyBytes and Content-Type for each part)
		mr := multipart.NewReader(strings.NewReader(body), params["boundary"])
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				// This err indicates that there are no more parts. So terminate the for loop.
				break
			}
			if err != nil {
				mykong.Err("Error from mr.NextPart():", err.Error())
				return false, fmt.Errorf("Failed to get Part:%v", mr)
			}

			fileName := part.FileName()
			mykong.Debug(fmt.Sprintf("=========== Start Check Part %v ===========", fileName))
			slurp, err := io.ReadAll(part)
			if err != nil {
				mykong.Err("Error from io.ReadAll(part):", err.Error())
				return false, fmt.Errorf("Failed to read bytes of file:%v", fileName)
			}

			// Correct representation of slurp depends upon the Content-Type value. If binary (e.g. application/octet-stream) then use %v else use %s (yes?).

			mykong.Debug("Part FileName:" + fmt.Sprintf("%v", fileName))
			mykong.Debug("Part Content-Type:" + fmt.Sprintf("%v", part.Header.Get("Content-Type")))
			// mykong.Debug("Part Body as bytes:" + fmt.Sprintf("%v", slurp))
			// mykong.Debug("Part Body as golang String:" + fmt.Sprintf("%q", slurp))

			// For each file, construct httpResp then get scan result.
			partIsOK, err := scanPart(mykong, scannerURL, part.FileName(), part.Header.Get("Content-Type"), slurp)
			if err != nil {
				mykong.Err("Error from scanPart(...):", err.Error())
				mykong.Debug(fmt.Sprintf("=========== Finish Check Part %v ===========", fileName))
				return false, fmt.Errorf("Failed to scan file:%v", fileName)
			} else if !partIsOK {
				mykong.Debug(fmt.Sprintf("=========== Finish Check Part %v ===========", fileName))
				return false, nil
			}

		}
	} else if len(body) > 0 {
		// The request is not multipart but it does have a body of some sort - scan the body.
		mykong.Debug("=========== Start Check non-multipart body ===========")
		partIsOK, err := scanPart(mykong, scannerURL, "", requestContentType, []byte(body))
		mykong.Debug("=========== Finish Check non-multipart body ===========")
		if err != nil {
			mykong.Err("Error from scanPart", err.Error())
			return false, fmt.Errorf("Failed to scan body")
		} else if !partIsOK {
			return false, nil
		}
	} else {
		mykong.Info(fmt.Sprintf("No body to check"))
	}
	mykong.Info(fmt.Sprintf("Finish checkBody"))
	return true, nil
}

func scanPart(mykong IKong, scannerURL string, fileName string, contentType string, bodyBytes []byte) (bool, error) {
	// TODO Consider using a single return of error

	mykong.Info(fmt.Sprintf("Start scanning body/part:'%v'", fileName))

	// TODO Control the DebugMode with config?
	ic.SetDebugMode(false)

	bodyByteBuffer := bytes.NewBuffer(bodyBytes)

	// Build an http.Response containing the body bytes that were in the request.
	h := make(http.Header, 0)
	h.Set("Content-Type", contentType)
	httpResp := &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          ioutil.NopCloser(bodyByteBuffer),
		ContentLength: int64(len(bodyBytes)),
		Header:        h,
	}

	/* making a icap request with OPTIONS method */
	optReq, err := ic.NewRequest(ic.MethodOPTIONS, scannerURL, nil, nil)

	if err != nil {
		mykong.Err("Err from ic.NewRequest:", err.Error())
		return false, err
	}

	/* making the icap client responsible for making the requests */
	client := &ic.Client{
		Timeout: 15 * time.Second,
	}

	mykong.Debug("Making the OPTIONS request call to:", scannerURL)

	// TODO Consider making this request cached or completely optional
	optResp, err := client.Do(optReq)

	if err != nil {
		mykong.Err("Err from client.Do(optReq):", err.Error())
		return false, err
	}

	mykong.Debug("optResp.PreviewBytes: ", optResp.PreviewBytes)

	/* making a icap request with RESPMOD method */
	req, err := ic.NewRequest(ic.MethodRESPMOD, scannerURL, nil, httpResp)

	if err != nil {
		mykong.Err("Err from ic.NewRequest(ic.MethodRESPMOD):", err.Error())
		return false, err
	}

	req.SetPreview(optResp.PreviewBytes) // setting the preview bytes obtained from the OPTIONS call

	/* making the RESPMOD request call */
	resp, err := client.Do(req)

	if err != nil {
		mykong.Err("Err from client.Do(req):", err.Error())
		return false, err
	}

	mykong.Debug("resp.StatusCode: ", resp.StatusCode)
	if resp.StatusCode == 204 {
		mykong.Info(fmt.Sprintf("Clean body/part:%v ", fileName))
		mykong.Debug(fmt.Sprintf("Finish scanning body/part:'%v'", fileName))
		return true, nil
	} else {
		isInfected := determineIsInfected(mykong, resp.Header)
		if isInfected {
			mykong.Warn(fmt.Sprintf("Infected body/part:%v ", fileName))
			mykong.Debug(fmt.Sprintf("Finish scanning body/part:'%v'", fileName))
		}
		return false, nil
	}
}

func determineIsInfected(mykong IKong, header http.Header) bool {
	infectionFound := header.Get("X-Infection-Found")
	mykong.Debug(fmt.Sprintf("determineIsInfected(): X-Infection-Found:'%s'", infectionFound))
	if len(infectionFound) > 0 {
		return true
	}

	virusID := header.Get("X-Virus-ID")
	mykong.Debug(fmt.Sprintf("determineIsInfected(): X-Virus-ID:'%s'", virusID))
	if len(virusID) > 0 && virusID != "no threats" {
		return true
	}

	fSecureScanResult := header.Get("X-FSecure-Scan-Result")
	mykong.Debug(fmt.Sprintf("determineIsInfected(): X-FSecure-Scan-Result:'%s'", fSecureScanResult))
	if fSecureScanResult == "infected" {
		return true
	}

	return false
}
