package db

// Additive tables keep existing backup schemas compatible.
type MaintenanceRule struct {
	TaskID string `gorm:"primaryKey" json:"taskId"`
	RepeatMonths int `json:"repeatMonths"`
	RepeatKm int `json:"repeatKm"`
	ReminderDays int `json:"reminderDays"`
	ReminderKm int `json:"reminderKm"`
	CompletedOdo *int `json:"completedOdo"`
	Notify bool `json:"notify"`
	NotificationStage int `json:"-"`
}
type RecurringCost struct {
	Base
	OwnerUserID string `gorm:"index;not null" json:"-"`
	VehicleID string `gorm:"index;not null" json:"vehicleId"`
	Title string `json:"title"`
	Category string `json:"category"`
	AmountCents int64 `json:"amountCents"`
	Currency string `json:"currency"`
	StartDate string `json:"startDate"`
	IntervalMonths int `json:"intervalMonths"`
	Installments int `json:"installments"` // zero = unlimited
	FinalCents *int64 `json:"finalCents"` // replaces last regular payment
	Generated int `json:"generated"`
	BookedCents int64 `json:"bookedCents"`
	Stopped bool `json:"stopped"`
	Notes string `json:"notes"`
}

type FinancingTerms struct {
 PlanID string `gorm:"primaryKey" json:"planId"`
 DownPaymentCents int64 `json:"downPaymentCents"`
 DownPaymentDate string `json:"downPaymentDate"`
 AlreadyRecorded bool `json:"alreadyRecorded"`
 PrincipalCents *int64 `json:"principalCents"`
 NominalBps *int `json:"nominalBps"`
 EffectiveBps *int `json:"effectiveBps"`
}
