package models

import (
	"encoding/json"
	"time"
)

type Place struct {
	ID          string      `json:"id" db:"id"`
	PlaceID     string      `json:"place_id" db:"place_id"`
	Name        string      `json:"name" db:"name"`
	Address     string      `json:"address" db:"address"`
	Coordinates Coordinates `json:"coordinates"`

	Categories []string `json:"categories"`

	Photos  []Photo  `json:"photos"`
	Reviews []Review `json:"reviews"`

	Rating      float32 `json:"rating"`
	UserRatings int     `json:"user_ratings"`
	PriceLevel  int     `json:"price_level"`
	Hours       string  `json:"hours"`
	Phone       string  `json:"phone"`
	Website     string  `json:"website"`

	UserNotes    string                 `json:"user_notes"`
	UserTags     []string               `json:"user_tags"`
	CustomFields map[string]interface{} `json:"custom_fields"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Photo struct {
	Reference string `json:"reference"`
	LocalPath string `json:"local_path"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
}

type Review struct {
	Author       string    `json:"author"`
	Rating       int       `json:"rating"`
	Text         string    `json:"text"`
	Time         time.Time `json:"time"`
	ProfilePhoto string    `json:"profile_photo"`
}

func (p *Place) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

func (p *Place) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

func (p *Place) HasTag(tag string) bool {
	for _, t := range p.UserTags {
		if t == tag {
			return true
		}
	}
	return false
}

func (p *Place) AddTag(tag string) {
	if !p.HasTag(tag) {
		p.UserTags = append(p.UserTags, tag)
	}
}

func (p *Place) RemoveTag(tag string) {
	for i, t := range p.UserTags {
		if t == tag {
			p.UserTags = append(p.UserTags[:i], p.UserTags[i+1:]...)
			break
		}
	}
}
