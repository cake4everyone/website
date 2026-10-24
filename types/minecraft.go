package types

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

const (
	MOJANG_API_URL = "https://api.mojang.com"

	MOJANG_API_PROFILE_BY_NAME = MOJANG_API_URL + "/users/profiles/minecraft/%s"

	MOJANG_API_ERROR_PROFILE_NOT_FOUND = "Couldn't find any profile"
)

type MinecraftProfile struct {
	UUID uuid.UUID `json:"id"`
	Name string
}

func GetMinecraftProfileByName(name string) (p MinecraftProfile, err error) {
	resp, err := http.Get(fmt.Sprintf(MOJANG_API_PROFILE_BY_NAME, name))
	if err != nil {
		return p, fmt.Errorf("failed to fetch minecraft profile by name: %w", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("failed to read the body of 'minecraft profile by name' respone: %+v", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		if strings.HasPrefix(string(body), MOJANG_API_ERROR_PROFILE_NOT_FOUND) {
			return p, fmt.Errorf("user %s not found", name)
		}
	}
	if resp.StatusCode != http.StatusOK {
		return p, fmt.Errorf("request 'minecraft profile by name' failed with %s: %s", resp.Status, body)
	}

	err = json.Unmarshal(body, &p)
	return p, err
}
