package cfgmodel

import validation "github.com/go-ozzo/ozzo-validation/v4"

// NotificationsConfig is the config for notifications sent by flipper to external services..
type NotificationsConfig struct {
	Enabled bool                       `json:"enabled"`
	Targets []NotificationTargetConfig `json:"targets"`
}

// Validate validates the notification config.
func (c NotificationsConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Enabled),
		validation.Field(&c.Targets),
	)
}

// NotificationTargetConfig is used for configuring a single target of notifications sent by flipper.
type NotificationTargetConfig struct {
	// Type of notification target, currently only "mattermost" is supported (which may be compatible with Slack).
	Type string `json:"type"`

	// URL of the Mattermost webhook.
	URL string `json:"url"`

	// Username to use when sending the notification.
	Username string `json:"username"`

	// Channel to send the notification to.
	// Note that in Mattermost this should be the channel slug, not the display name.
	Channel string `json:"channel"`

	// IconEmoji to use when sending the notification.
	// Defaults to ":dolphin:".
	IconEmoji string `json:"icon_url"`
}

// Validate validates the notification config.
func (c NotificationTargetConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Type, validation.Required, validation.In("mattermost")),
		validation.Field(&c.URL, validation.Required),
	)
}

// IconEmojiOrDefault returns the icon emoji or the default value.
func (c NotificationTargetConfig) IconEmojiOrDefault() string {
	if c.IconEmoji == "" {
		return ":dolphin:"
	}
	return c.IconEmoji
}
