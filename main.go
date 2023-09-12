package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Endpoint
	r.POST("/calculate-sla", CalculateSLA)

	r.Run(":8080")
}

// RequestPayload
type RequestPayload struct {
	CreateTime time.Time `json:"create_time"`
	SLARef     string    `json:"sla_ref"`
}

// Response
type ResponsePayload struct {
	SLA50Percentage  float64          `json:"sla_50_percentage"`
	SLA75Percentage  float64          `json:"sla_75_percentage"`
	SLA100Percentage float64          `json:"sla_100_percentage"`
	Details          map[string]int64 `json:"details"`
}

// handler untuk menghitung SLA
func CalculateSLA(c *gin.Context) {
	var request RequestPayload
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slaRules := map[string]int{
		"A": 24,
		"B": 72,
		"C": 144,
	}

	slaHours := slaRules[request.SLARef]
	endTime := request.CreateTime.Add(time.Duration(slaHours) * time.Hour)

	// Hitung detail
	details := calculateSLADetails(request.CreateTime, endTime)

	// Hitung persentase
	sla50Percentage := calculateSLAPercentage(request.CreateTime, endTime, 0.5)
	sla75Percentage := calculateSLAPercentage(request.CreateTime, endTime, 0.75)
	sla100Percentage := calculateSLAPercentage(request.CreateTime, endTime, 1.0)

	response := ResponsePayload{
		SLA50Percentage:  sla50Percentage,
		SLA75Percentage:  sla75Percentage,
		SLA100Percentage: sla100Percentage,
		Details:          details,
	}

	c.JSON(http.StatusOK, response)
}

func calculateSLADetails(startTime, endTime time.Time) map[string]int64 {
	details := make(map[string]int64)

	// Hari kerja adalah Senin hingga Jumat
	for day := startTime; day.Before(endTime); day = day.Add(24 * time.Hour) {
		if day.Weekday() == time.Saturday || day.Weekday() == time.Sunday {
			// Hari Sabtu dan Minggu tidak termasuk dalam hari kerja
			continue
		}

		// Hitung jam kerja dalam sehari
		workStartTime := time.Date(day.Year(), day.Month(), day.Day(), 9, 0, 0, 0, time.UTC)
		workEndTime := time.Date(day.Year(), day.Month(), day.Day(), 18, 0, 0, 0, time.UTC)
		breakStartTime := time.Date(day.Year(), day.Month(), day.Day(), 12, 0, 0, 0, time.UTC)
		breakEndTime := time.Date(day.Year(), day.Month(), day.Day(), 13, 0, 0, 0, time.UTC)

		workHours := calculateWorkHoursInDay(startTime, endTime, workStartTime, workEndTime, breakStartTime, breakEndTime)

		// Simpan detail perhitungan SLA untuk hari ini
		details[day.Format("02_Jan_06")] = workHours
	}

	return details
}

func calculateWorkHoursInDay(startTime, endTime, workStartTime, workEndTime, breakStartTime, breakEndTime time.Time) int64 {
	// Batasi waktu kerja dengan waktu mulai dan waktu selesai
	if startTime.Before(workStartTime) {
		startTime = workStartTime
	}
	if endTime.After(workEndTime) {
		endTime = workEndTime
	}

	// Hitung waktu kerja dalam jam
	workHours := int64(endTime.Sub(startTime).Hours())

	// Mengurangkan waktu istirahat dari waktu kerja
	if startTime.Before(breakStartTime) && endTime.After(breakEndTime) {
		workHours -= int64(breakEndTime.Sub(breakStartTime).Hours())
	}

	return workHours
}

func calculateSLAPercentage(startTime, endTime time.Time, targetPercentage float64) float64 {
	daysInRange := 0
	targetDays := int(float64(endTime.Sub(startTime).Hours())/24.0*targetPercentage) + 1
	currentDay := startTime
	for currentDay.Before(endTime) {
		if currentDay.Weekday() != time.Saturday && currentDay.Weekday() != time.Sunday {
			daysInRange++
		}
		currentDay = currentDay.Add(24 * time.Hour)
		if daysInRange >= targetDays {
			break
		}
	}

	return float64(daysInRange) / float64(targetDays) * 100.0
}
