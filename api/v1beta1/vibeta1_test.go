package v1beta1

import (
	"fmt"
	"testing"
)

type app struct {
	Group   string
	Version string
	Kind    string
}

func TestCreatePostHandler(t *testing.T) {

	var apps []app
	apps = []app{
		*NewApp(),
	}
	fmt.Println(apps)
}

func NewApp() *app {
	return &app{
		Group:   "1",
		Version: "2",
		Kind:    "3",
	}
}
