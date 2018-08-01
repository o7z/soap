package soap

import (
	"net/http"
	"io"
	"encoding/xml"
	"fmt"
	"strings"
	"bytes"
	"io/ioutil"
)

type Helper struct {
	serviceURL string
	namespace  string
}

func NewHelper(serviceURL, namespace string) (*Helper) {
	return &Helper{
		serviceURL: serviceURL,
		namespace:  namespace}
}

type SOAPEnvelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    SOAPBody `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
}

type SOAPBody struct {
	Body interface{}
}

type Node struct {
	Parent             *Node
	Namespace          string
	Name               string
	Value              string
	Children           []*Node
	namespaceValueKeys map[string]string
}

func (h *Helper) Request(body interface{}) (string, []byte, error) {
	if req, err := http.NewRequest("POST", h.serviceURL, h.getRequestXMLBody(body)); err != nil {
		return "", nil, err
	} else {
		cli := new(http.Client)
		if resp, err := cli.Do(req); err != nil {
			return "", nil, err
		} else {
			defer resp.Body.Close()
			respName, respContentBuf := decodeResponse(resp.Body)
			return respName, respContentBuf, nil
		}
	}
}

func (h *Helper) Request2(body interface{}, respContent interface{}) (string, error) {
	if req, err := http.NewRequest("POST", h.serviceURL, h.getRequestXMLBody(body)); err != nil {
		return "", err
	} else {
		cli := new(http.Client)
		if resp, err := cli.Do(req); err != nil {
			return "", err
		} else {
			defer resp.Body.Close()
			respName, respContentBuf := decodeResponse(resp.Body)
			if err := xml.Unmarshal(respContentBuf, respContent); err != nil {
				return respName, err
			}
			return respName, nil
		}
	}
}

func (h *Helper) RequestTest(body interface{}) (string, error) {
	if req, err := http.NewRequest("POST", h.serviceURL, h.getRequestXMLBody(body)); err != nil {
		return "", err
	} else {
		cli := new(http.Client)
		if resp, err := cli.Do(req); err != nil {
			return "", err
		} else {
			defer resp.Body.Close()
			if buf, err := ioutil.ReadAll(resp.Body); err != nil {
				return "", err
			} else {
				return string(buf), nil
			}
		}
	}
}

func (h *Helper) getRequestXMLBody(body interface{}) (io.Reader) {
	envelope := createSoapEnvelope(body)
	buffer := &bytes.Buffer{}
	encoder := xml.NewEncoder(buffer)
	encoder.Encode(envelope)
	return buffer
}

func createSoapEnvelope(body interface{}) *SOAPEnvelope {
	soapEnvelope := &SOAPEnvelope{
		Body: SOAPBody{Body: body}}
	return soapEnvelope
}

/*
----
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
<soap:Body>
<ns1:getPortalRequestResponse xmlns:ns1="http://www.bnet.cn/v3.0">
<getPortalRequestResponse>
&lt;?xml version=&quot;1.0&quot; encoding=&quot;UTF-8&quot; standalone=&quot;yes&quot;?&gt;
&lt;Package&gt;&lt;StreamingNo&gt;20180712161944641269&lt;/StreamingNo&gt;&lt;OPFlag&gt;0101&lt;/OPFlag&gt;&lt;TimeStamp&gt;20180712161944&lt;/TimeStamp&gt;&lt;ProductID&gt;GX9900058&lt;/ProductID&gt;&lt;BizID&gt;11111111119058&lt;/BizID&gt;&lt;AreaCode&gt;774&lt;/AreaCode&gt;&lt;CustID&gt;825838&lt;/CustID&gt;&lt;CustAccount&gt;test0721&lt;/CustAccount&gt;&lt;CustName&gt;ceshi&lt;/CustName&gt;&lt;/Package&gt;
</getPortalRequestResponse></ns1:getPortalRequestResponse></soap:Body></soap:Envelope>
----
*/

func decodeResponse(bodyReader io.Reader) (respName string, respContentBuf []byte) {
	decoder := xml.NewDecoder(bodyReader)
	toDecodeBody := false
	toDecodeBodyContent := true
	for token, err := decoder.Token(); err == nil; token, err = decoder.Token() {
		switch ele := token.(type) {
		case xml.StartElement:
			//name := getNameStr(ele.Name)
			//attrStr := getAttrsStr(ele.Attr)
			//if attrStr == "" {
			//	fmt.Printf("<%s>\n\r", name)
			//} else {
			//	fmt.Printf("<%s %s>\n\r", name, attrStr)
			//}
			if toDecodeBody {
				respName = ele.Name.Local
			}
			if ele.Name.Local == "Body" {
				toDecodeBody = true
			}
		case xml.EndElement:
			//name := getNameStr(ele.Name)
		case xml.CharData:
			if toDecodeBodyContent {
				respContentBuf = ele
				return
			}
		default:
			fmt.Printf("%s(%T)\n\r", ele, ele)
		}
	}
	return
}

func (n Node) TryGetNamespaceKey(v string) string {
	for {
		if k, ex := n.namespaceValueKeys[v]; ex {
			return k
		} else if n.Parent != nil {
			return n.Parent.TryGetNamespaceKey(v)
		} else {
			return ""
		}
	}
}

func appendNS(m map[string]string, k, v string) {
	m[k] = v
}

func getAttrsStr(attrs []xml.Attr) string {
	attrStrs := []string{}
	for _, attr := range attrs {
		attrStrs = append(attrStrs, getAttrStr(attr))
	}
	return strings.Join(attrStrs, " ")
}

func getAttrStr(attr xml.Attr) string {
	return fmt.Sprintf("%s=\"%s\"", getNameStr(attr.Name), attr.Value)
}

func getNameStr(name xml.Name) (string) {
	if name.Space == "" {
		return name.Local
	} else {
		return fmt.Sprintf("%s:%s", name.Space, name.Local)
	}
}
