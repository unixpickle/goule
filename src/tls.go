package goule

type CertificateInfo struct {
	CertificatePath string   `json:"certificate_path"`
	KeyPath         string   `json:"key_path"`
	AuthorityPaths  []string `json:"authority_paths"`
}

type TLSInfo struct {
	Named   map[string]CertificateInfo `json:"named_certificates"`
	Default CertificateInfo            `json:"default_certificates"`
}
