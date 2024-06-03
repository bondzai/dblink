package domain

import "github.com/google/uuid"

type DriverDTO struct {
	Location     DriverLocation     `json:"location"`
	LoginSession DriverLoginSession `json:"loginSession"`
	Type         DriverType         `json:"type"`
	Job          *uuid.UUID         `json:"job"`
}

type DriverLocation struct {
	Lat  string `json:"lat"`
	Long string `json:"long"`
}

type DriverLoginSession struct {
	DeviceID string `json:"deviceId"`
}

type DriverType struct {
	CompanyApproveStatus int       `json:"companyApproveStatus"`
	JobAcceptStatus      *int      `json:"jobAcceptStatus"`
	IsInternalCompany    bool      `json:"isInternalCompany"`
	VehicleTypeID        uuid.UUID `json:"vehicleTypeId"`
}
