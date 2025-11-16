package models

type Environment struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	PythonPath string `json:"python_path"`
	IsValid    bool   `json:"is_valid"`
}
