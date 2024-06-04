package service

import (
	"log/slog"

	"github.com/bondzai/dblink/internal/domain"
	"github.com/bondzai/dblink/internal/repository"
)

type DriverWsService interface {
	GetLatestData(driverID string) *domain.DriverWsDto
	ProcessUpdate(driverID string, updateData map[string]interface{}) *domain.DriverWsDto
}

type driverWsService struct {
	repo repository.RedisRepository
}

func NewDriverWsService(repo repository.RedisRepository) DriverWsService {
	return &driverWsService{
		repo: repo,
	}
}

func (s *driverWsService) GetLatestData(driverID string) *domain.DriverWsDto {
	driver, err := s.repo.GetDriver(driverID)

	if err != nil {
		slog.Error("Error getting data in Redis", err)
		return &domain.DriverWsDto{}
	}

	if driver.Id == "" {
		defaultDriver := s.getDefaultData(driverID)
		if err := s.repo.SaveDriver(defaultDriver); err != nil {
			slog.Error("Error setting default driver data in Redis", err)
			return &domain.DriverWsDto{}
		}
		return defaultDriver
	}

	return driver
}

func (s *driverWsService) ProcessUpdate(driverID string, updateData map[string]interface{}) *domain.DriverWsDto {
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
			// todo: save data to db
			driver.LoginSession.DeviceID = deviceId
		}
	}

	if driverType, ok := updateData["type"].(map[string]interface{}); ok {
		if isInternalCompany, ok := driverType["isInternalCompany"].(bool); ok {
			// todo: save data to db
			driver.Type.IsInternalCompany = isInternalCompany
		}
	}

	if err := s.repo.SaveDriver(driver); err != nil {
		slog.Error("Error setting data in Redis", err)
	}

	return driver
}

// todo: query data from db and other services instead of mock.
func (s *driverWsService) getDefaultData(driverID string) *domain.DriverWsDto {
	return &domain.DriverWsDto{
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
			VehicleTypeID:        nil,
		},
		JobId: nil,
	}
}
