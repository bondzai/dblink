package repositories

import "github.com/bondzai/dblink/internal/models"

type mockRepo struct{}

func NewMockRepo() *mockRepo {
	return &mockRepo{}
}

func (m *mockRepo) MockUsers() []models.User {
	return []models.User{
		{
			UserID:   1,
			UserName: "John",
			Location: models.Location{
				Lat:  1.1,
				Long: 1.1,
			},
		},
		{
			UserID:   2,
			UserName: "Jane",
			Location: models.Location{
				Lat:  2.2,
				Long: 2.2,
			},
		},
		{
			UserID:   3,
			UserName: "Doe",
			Location: models.Location{
				Lat:  3.3,
				Long: 3.3,
			},
		},
	}
}
