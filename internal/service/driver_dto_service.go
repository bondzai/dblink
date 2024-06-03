package service

import (
	"log"

	"github.com/bondzai/dblink/internal/domain"
	"github.com/bondzai/dblink/internal/repository"
	"github.com/google/uuid"
)

type DriverService struct {
	repo *repository.RedisRepository
}

func NewDriverService(repo *repository.RedisRepository) *DriverService {
	return &DriverService{
		repo: repo,
	}
}

func (s *DriverService) GetLatestData(driverID string) domain.DriverDTO {
	driver, err := s.repo.GetDriver(driverID)

	if err != nil {
		log.Printf("Error getting data from Redis: %v", err)
		return domain.DriverDTO{}
	}

	if driver.Id == "" {
		defaultDriver := s.getDefaultData(driverID)
		if err := s.repo.SaveDriver(defaultDriver); err != nil {
			log.Printf("Error setting default driver data in Redis: %v", err)
			return domain.DriverDTO{}
		}
		return defaultDriver
	}

	return driver
}

func (s *DriverService) ProcessUpdate(driverID string, updateData map[string]interface{}) *domain.DriverDTO {
	driver := s.GetLatestData(driverID)

	if loc, ok := updateData["location"].(map[string]interface{}); ok {
		if lat, ok := loc["lat"].(string); ok {
			driver.Location.Lat = lat
		}
		if long, ok := loc["long"].(string); ok {
			driver.Location.Long = long
		}
	}

	if session, ok := updateData["loginSession"].(map[string]interface{}); ok {
		if deviceId, ok := session["deviceId"].(string); ok {
			driver.LoginSession.DeviceID = deviceId
		}
	}

	if driverType, ok := updateData["type"].(map[string]interface{}); ok {
		if isInternalCompany, ok := driverType["isInternalCompany"].(bool); ok {
			driver.Type.IsInternalCompany = isInternalCompany
		}
	}

	if err := s.repo.SaveDriver(driver); err != nil {
		log.Printf("Error setting data in Redis: %v", err)
	}

	return &driver
}

func (s *DriverService) getDefaultData(driverID string) domain.DriverDTO {
	return domain.DriverDTO{
		Id: driverID,
		Location: domain.DriverLocation{
			Lat:  "0",
			Long: "0",
		},
		LoginSession: domain.DriverLoginSession{
			DeviceID: "default-device-id",
		},
		Type: domain.DriverType{
			CompanyApproveStatus: 0,
			JobAcceptStatus:      nil,
			IsInternalCompany:    false,
			VehicleTypeID:        uuid.Nil,
		},
		Job: nil,
	}
}
