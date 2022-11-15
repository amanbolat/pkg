package email

type Email struct {
	From     string `json:",omitempty"`
	To       string `json:",omitempty"`
	Cc       string `json:",omitempty"`
	Bcc      string `json:",omitempty"`
	Subject  string `json:",omitempty"`
	HtmlBody string `json:",omitempty"`
	TextBody string `json:",omitempty"`
}

type TemplatedEmail struct {
	TemplateId   int64                  `json:",omitempty"`
	TemplateData map[string]interface{} `json:",omitempty"`
	From         string                 `json:",omitempty"`
	To           string                 `json:",omitempty"`
	Cc           string                 `json:",omitempty"`
	Bcc          string                 `json:",omitempty"`
}
