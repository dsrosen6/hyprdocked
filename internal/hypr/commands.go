package hypr

import "github.com/dsrosen6/hyprlaptop/internal/models"

func (c *Client) ListMonitors() ([]models.Monitor, error) {
	var m []models.Monitor
	if err := c.RunCommandWithUnmarshal([]string{"monitors"}, &m); err != nil {
		return nil, err
	}

	return m, nil
}
