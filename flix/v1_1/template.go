package v1_1

type Template struct {
	rawJson []byte
}

func NewFromJson(rawJson []byte) (Template, error) {
	return Template{
		rawJson: rawJson,
	}, nil
}

func (t Template) MarshalJSON() ([]byte, error) {
	return t.rawJson, nil
}
