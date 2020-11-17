package extracaocampo

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
)

type Node struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:"-"`
	Content []byte     `xml:",innerxml"`
	Nodes   []Node     `xml:",any"`
	Parent  string
}

//Extrair - Executa a extração de campos
func Extrair(payload string) (map[string]string, error) {
	campos := make(map[string]string)
	var err error
	if len(payload) == 0 {
		panic("Payload vazio para a ExtracaoCampos")
	}
	if strings.HasPrefix(payload, "<") {
		err = extrairCamposXML(payload, "", 0, campos)
	} else {
		err = extrairCamposjson(payload, "", 0, campos)
	}

	return campos, err

}

// extrairCampos - retorna os campos no padrão
func extrairCamposjson(dados string, parentKey string, index int, campos map[string]string) error {
	var modelo = make(map[string]interface{})

	bytes := []byte(dados)
	err := json.Unmarshal(bytes, &modelo)
	if err != nil {
		return err
	}

	for fieldname, element := range modelo {
		var ValorTag string
		if len(parentKey) > 0 {
			ValorTag = parentKey + "." + fieldname
			if index > 0 {
				ValorTag = ValorTag + fmt.Sprintf(" %v", index)
			}

		} else {
			ValorTag = fieldname
		}

		switch element.(type) {
		case []interface{}:
			jsondata, _ := json.Marshal(element)
			campos[ValorTag] = string(jsondata)
			for _, elementChild := range element.([]interface{}) {
				datachild, _ := json.Marshal(elementChild)

				extrairCamposjson(string(datachild), ValorTag, index, campos)
				index++
			}
		case map[string]interface{}:
			for childName, elementChild := range element.(map[string]interface{}) {
				childTag := ValorTag + "." + childName
				campos[childTag] = fmt.Sprintf("%v", elementChild)
			}

		default:
			campos[ValorTag] = fmt.Sprintf("%v", element)
		}
	}
	return nil
}

// extrairCampos - retorna os campos no padrão
func extrairCamposXML(dados string, parentKey string, index int, campos map[string]string) error {
	data := []byte(dados)
	buf := bytes.NewBuffer(data)
	dec := xml.NewDecoder(buf)
	var modelo Node

	err := dec.Decode(&modelo)
	if err != nil {
		return err
	}

	for index, element := range []Node{modelo} {

		var ValorTag string
		if len(parentKey) > 0 {
			ValorTag = parentKey + "/" + element.XMLName.Local
			if index > 0 {
				ValorTag = ValorTag + fmt.Sprintf(" %v", index)
			}

		} else {
			ValorTag = element.XMLName.Local
		}

		if len(element.Nodes) > 0 {
			for _, elementChild := range element.Nodes {
				elementChild.Parent = ValorTag
				getCamposXML(elementChild, ValorTag, index, campos)
			}

		} else {
			campos[ValorTag] = fmt.Sprintf(string(element.Content))
			if len(element.Parent) > 0 {
				campos[element.Parent] += fmt.Sprintf(string(element.Content))
			}
		}

	}
	return nil
}

// extrairCampos - retorna os campos no padrão
func getCamposXML(element Node, parentKey string, index int, campos map[string]string) error {

	var ValorTag string
	if len(parentKey) > 0 {
		ValorTag = parentKey + "/" + element.XMLName.Local
		if index > 0 {
			ValorTag = ValorTag + fmt.Sprintf(" %v", index)
		}

	} else {
		ValorTag = element.XMLName.Local
	}

	if len(element.Nodes) > 0 {
		for _, elementChild := range element.Nodes {
			elementChild.Parent = ValorTag
			getCamposXML(elementChild, ValorTag, index, campos)
		}
		if len(element.Parent) > 0 {
			campos[element.Parent] = campos[ValorTag]
		}
	} else {

		campos[ValorTag] = fmt.Sprintf(string(element.Content))
		if len(element.Attrs) > 0 {
			for _, attr := range element.Attrs {
				valorNameAtributo := ValorTag + "/" + attr.Name.Local + ":" + attr.Value
				campos[attr.Value] = fmt.Sprintf(string(element.Content))
				campos[valorNameAtributo] = fmt.Sprintf(string(element.Content))
			}
		}

		if len(element.Parent) > 0 {
			campos[element.Parent] += fmt.Sprintf(string(element.Content))
		}
	}

	return nil
}

func (n *Node) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	n.Attrs = start.Attr
	type node Node

	return d.DecodeElement((*node)(n), &start)
}
