package goworkouts

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"time"
	"fmt"

	"github.com/google/uuid"

	//"github.com/muktihari/fit"
	"github.com/tormoder/fit"
	yaml "gopkg.in/yaml.v2"
)

// Mapping of string-based sport names to fit.Sport values
var sportMapping = map[string]fit.Sport{
	"running":   fit.SportRunning,
	"cycling":   fit.SportCycling,
	"swimming":  fit.SportSwimming,
	"walking":   fit.SportWalking,
	"generic":   fit.SportGeneric,
	"rowing":    fit.SportRowing,
	"hiking":    fit.SportHiking,
	"multi":     fit.SportMultisport,
}

// WorkoutStep is the container of a Workout Step
type WorkoutStep struct {
	MessageIndex          fit.MessageIndex `json:"stepId" yaml:"stepId"`
	WktStepName           string           `json:"wkt_step_name" yaml:"wkt_step_name"`
	DurationType          string           `json:"durationType" yaml:"durationType"`
	DurationValue         uint32           `json:"durationValue" yaml:"durationValue"`
	TargetType            string           `json:"targetType" yaml:"targetType"`
	TargetValue           uint32           `json:"targetValue" yaml:"targetValue"`
	CustomTargetValueLow  uint32           `json:"targetValueLow" yaml:"targetValueLow"`
	CustomTargetValueHigh uint32           `json:"targetValueHigh" yaml:"targetValueHigh"`
	Intensity             string           `json:"intensity" yaml:"intensity"`
	Notes                 string           `json:"description" yaml:"description"`
	// Type                  string           `json:"type"`
}

// NewWorkoutStep creates new workout step
func newWorkoutStep() WorkoutStep {
	newstep := WorkoutStep{}
	newstep.DurationValue = 0
	newstep.TargetValue = 0
	newstep.CustomTargetValueLow = 0
	newstep.CustomTargetValueHigh = 0
	return newstep
}

// Workout is a Workout
type Workout struct {
	Filename    string        `json:"filename" yaml:"filename"`
	Name        string        `json:"workoutName" yaml:"workoutName"`
	Steps       []WorkoutStep `json:"steps" yaml:"steps"`
	Sport       string        `json:"sport" yaml:"sport"`
	Description string        `json:"description" yaml:"description"`
	// WorkoutID uint64         `json:"WorkoutId"`
	// OwnerID   uint64         `json:"ownerId"`
}

// ToJSON export to JSON
func (w *Workout) ToJSON() ([]byte, error) {
	return json.Marshal(w)
}

// FitPowerConversion converts Power to zone or value
func FitPowerConversion(step WorkoutStep) (string, error) {
	pwr := step.TargetValue
	pwrlow := step.CustomTargetValueLow
	pwrhigh := step.CustomTargetValueHigh
	// pwr is a zone setting
	if pwr > 0 {
		return fmt.Sprintf("Z%v", pwr), nil
	}
	if pwrhigh <= 1000 {
		return fmt.Sprintf("%v-%v%%", pwrlow, pwrhigh), nil
	}
	return fmt.Sprintf("%v-%vW", pwrlow-1000, pwrhigh-1000), nil
	
}

// FitHRConversion converts Power to zone or value
func FitHRConversion(step WorkoutStep) (string, error) {
	hr := step.TargetValue
	hrlow := step.CustomTargetValueLow
	hrhigh := step.CustomTargetValueHigh
	if hr > 0 {
		return fmt.Sprintf("Z%v HR", hr), nil
	}
	if hrhigh <= 100 {
		return fmt.Sprintf("%v-%v%% HR", hrlow, hrhigh), nil
	}
	return fmt.Sprintf("%v-%v HR", hrlow-100, hrhigh-100), nil
	
}

func GetStepByIndex(w Workout, idx fit.MessageIndex) (WorkoutStep, error) {
	for _, step := range w.Steps {
		if step.MessageIndex == idx {
			return step, nil
		}
	}
	return WorkoutStep{}, errors.New("Step not found")
}

func AddRepeats(stepslist []string, idxlist []fit.MessageIndex, idx fit.MessageIndex, nr_repeats uint32) ([]string) {
	for i, _ := range(stepslist) {
		if idxlist[i] == idx {
			stepslist[i] = fmt.Sprintf("\n%vx\n", nr_repeats)+stepslist[i]
		}
	}
	return stepslist
}

// ToIntervals exports to intervals.icu workout description language
// Each step is turned into a string like "- 10m @ 200W Comment"
func (w *Workout) ToIntervals() (string, error) {
	var err error
	var stepstext string
	var stepslist []string
	var idxlist []fit.MessageIndex
	for _, step := range w.Steps {
		idxlist = append(idxlist, step.MessageIndex)
		var buffer bytes.Buffer
		var prefix, duration, target, name, notes, intensity string
		if step.DurationType == "RepeatUntilStepsCmplt" {
			nr_repeats := step.TargetValue
			idx := fit.MessageIndex(step.DurationValue)
			stepslist = AddRepeats(stepslist, idxlist, idx, nr_repeats)
			stepslist = append(stepslist, "\n")
		} else{
			if step.DurationType == "Time" {
				seconds := float64(step.DurationValue)/1000.
					duration = fmt.Sprintf("%vs", seconds)
			}
			if step.DurationType == "Distance" {
				meters := float64(step.DurationValue)/1.e5
					duration = fmt.Sprintf("%vkm", meters)
			}
			if step.TargetType == "Power" || step.TargetType == "PowerLap" {
				target, err = FitPowerConversion(step)
				if err != nil {
					target = ""
				}
			}
			if step.TargetType == "HeartRate" || step.TargetType == "HeartRateLap" {
				target, err = FitHRConversion(step)
				if err != nil {
					target = ""
				}
			}
			if step.TargetType == "Cadence" {
				spm := step.TargetValue
				if spm > 0 {
					target = fmt.Sprintf("%vrpm", spm)
				} else {
					spmlow := step.CustomTargetValueLow
					spmhigh := step.CustomTargetValueHigh
					target = fmt.Sprintf("%v-%vrpm", spmlow, spmhigh)
				}
			}
			if step.Intensity == "Warmup" {
				prefix = "\nWarmup\n"
			}
			if step.Intensity == "Cooldown" {
				prefix = "\nCooldown\n"
			}
			// Speed
			// 
			name = step.WktStepName
			notes = step.Notes
			intensity = step.Intensity
			buffer.WriteString(fmt.Sprintf("%v- %v %v %v %v %v\n", prefix, duration, target, intensity, name, notes))
			stepslist = append(stepslist, buffer.String())
		}
	}

	var buffer bytes.Buffer
	for _, txt := range stepslist {
		buffer.WriteString(txt)
	}

	stepstext = buffer.String()
	
	return stepstext, nil
}

// ToYAML export to YAML
func (w *Workout) ToYAML() ([]byte, error) {
	return yaml.Marshal(w)
}

// FromJSON returns workout from json string
func FromJSON(s string) (Workout, error) {
	var w Workout
	err := json.Unmarshal([]byte(s), &w)
	return w, err
}

// FromYAML returns workout from YAML
func FromYAML(s string) (Workout, error) {
	var w Workout
	err := yaml.Unmarshal([]byte(s), &w)
	return w, err
}

// TrainingDay is a training day in the training plan
type TrainingDay struct {
	Order    uint32    `json:"order" yaml:"order"` // how manieth calendar day
	Workouts []Workout `json:"workouts" yaml:"workouts"`
}

// TrainingPlan is a training plan
type TrainingPlan struct {
	ID           uuid.UUID     `json:"ID" yaml:"ID"`
	Filename     string        `json:"filename" yaml:"filename"`
	Name         string        `json:"name" yaml:"name"`
	TrainingDays []TrainingDay `json:"trainingDays" yaml:"trainingDays"`
	Duration     uint32        `json:"duration" yaml:"duration"` // in number of calendar days
	Description  string        `json:"description" yaml:"description"`
}

// NewTrainingPlan type for training plans that do not exist yet
type NewTrainingPlan struct {
	Filename     string        `json:"filename" yaml:"filename"`
	Name         string        `json:"name" yaml:"name"`
	TrainingDays []TrainingDay `json:"trainingDays" yaml:"trainingDays"`
	Duration     uint32        `json:"duration" yaml:"duration"` // in number of calendar days
	Description  string        `json:"description" yaml:"description"`
}

// ListOfPlans storing list of plans
type ListOfPlans struct {
	Plans []TrainingPlan `json:"plans" yaml:"plans"`
}

var targetTypes = map[string]fit.WktStepTarget{
	"Speed":        fit.WktStepTargetSpeed,        //        WktStepTarget = 0
	"HeartRate":    fit.WktStepTargetHeartRate,    //    WktStepTarget = 1
	"Open":         fit.WktStepTargetOpen,         //         WktStepTarget = 2
	"Cadence":      fit.WktStepTargetCadence,      //      WktStepTarget = 3
	"Power":        fit.WktStepTargetPower,        //        WktStepTarget = 4
	"Grade":        fit.WktStepTargetGrade,        //        WktStepTarget = 5
	"Resistance":   fit.WktStepTargetResistance,   //   WktStepTarget = 6
	"Power3s":      fit.WktStepTargetPower3s,      //   WktStepTarget = 7
	"Power10s":     fit.WktStepTargetPower10s,     //    WktStepTarget = 8
	"Power30s":     fit.WktStepTargetPower30s,     //   WktStepTarget = 9
	"PowerLap":     fit.WktStepTargetPowerLap,     //   WktStepTarget = 10
	"SwimStroke":   fit.WktStepTargetSwimStroke,   //  WktStepTarget = 11
	"SpeedLap":     fit.WktStepTargetSpeedLap,     // WktStepTarget = 12
	"HeartRateLap": fit.WktStepTargetHeartRateLap, // WktStepTarget = 13
}

var intensityTypes = map[string]fit.Intensity{
	"Active":   fit.IntensityActive,   //   Intensity = 0
	"Rest":     fit.IntensityRest,     //  Intensity = 1
	"Warmup":   fit.IntensityWarmup,   // Intensity = 2
	"Cooldown": fit.IntensityCooldown, // Intensity = 3
	"Recovery": fit.IntensityRecovery, // Intensity = 4
	"Interval": fit.IntensityInterval, // Intensity = 5
	"Other": fit.IntensityOther, // Intensity = 6
	//IntensityInvalid  Intensity = 0xFF
}

var durationTypes = map[string]fit.WktStepDuration{
	"Time":                               fit.WktStepDurationTime,
	"Distance":                           fit.WktStepDurationDistance,
	"HrLessThan":                         fit.WktStepDurationHrLessThan,                      //                         // WktStepDuration = 2
	"HrGreaterThan":                      fit.WktStepDurationHrGreaterThan,                   //                      // WktStepDuration = 3
	"Calories":                           fit.WktStepDurationCalories,                        //                           // WktStepDuration = 4
	"Open":                               fit.WktStepDurationOpen,                            //                               // WktStepDuration = 5
	"RepeatUntilStepsCmplt":              fit.WktStepDurationRepeatUntilStepsCmplt,           // WktStepDuration = 6
	"RepeatUntilTime":                    fit.WktStepDurationRepeatUntilTime,                 // WktStepDuration = 7
	"RepeatUntilDistance":                fit.WktStepDurationRepeatUntilDistance,             // WktStepDuration = 8
	"RepeatUntilCalories":                fit.WktStepDurationRepeatUntilHrLessThan,           // WktStepDuration = 9
	"RepeatUntilHrLessThan":              fit.WktStepDurationRepeatUntilHrLessThan,           // WktStepDuration = 10
	"RepeatUntilHrGreaterThan":           fit.WktStepDurationRepeatUntilHrGreaterThan,        // WktStepDuration = 11
	"RepeatUntilPowerLessThan":           fit.WktStepDurationRepeatUntilPowerLessThan,        // WktStepDuration = 12
	"RepeatUntilPowerGreaterThan":        fit.WktStepDurationRepeatUntilPowerGreaterThan,     // WktStepDuration = 13
	"PowerLessThan":                      fit.WktStepDurationPowerLessThan,                   // WktStepDuration = 14
	"PowerGreaterThan":                   fit.WktStepDurationPowerGreaterThan,                // WktStepDuration = 15
	"TrainingPeaksTss":                   fit.WktStepDurationTrainingPeaksTss,                // WktStepDuration = 16
	"RepeatUntilPowerLastLapLessThan":    fit.WktStepDurationRepeatUntilPowerLastLapLessThan, // WktStepDuration = 17
	"RepeatUntilMaxPowerLastLapLessThan": fit.WktStepDurationRepeatUntilPowerLastLapLessThan, // WktStepDuration = 18
	"Power3sLessThan":                    fit.WktStepDurationPower3sLessThan,                 // WktStepDuration = 19
	"Power10sLessThan":                   fit.WktStepDurationPower10sLessThan,                // WktStepDuration = 20
	"Power30sLessThan":                   fit.WktStepDurationPower30sLessThan,                // WktStepDuration = 21
	"Power3sGreaterThan":                 fit.WktStepDurationPower3sGreaterThan,              // WktStepDuration = 22
	"Power10sGreaterThan":                fit.WktStepDurationPower10sGreaterThan,             // WktStepDuration = 23
	"Power30sGreaterThan":                fit.WktStepDurationPower30sGreaterThan,             // WktStepDuration = 24
	"PowerLapLessThan":                   fit.WktStepDurationPowerLapLessThan,                // WktStepDuration = 25
	"PowerLapGreaterThan":                fit.WktStepDurationPowerLapGreaterThan,             // WktStepDuration = 26
	"RepeatUntilTrainingPeaksTss":        fit.WktStepDurationRepeatUntilTrainingPeaksTss,     // WktStepDuration = 27
	"RepetitionTime":                     fit.WktStepDurationRepetitionTime,                  // WktStepDuration = 28
	"Reps":                               fit.WktStepDurationReps,                            // WktStepDuration = 29
	"TimeOnly":                           fit.WktStepDurationTimeOnly,                        // WktStepDuration = 31
	//"Invalid":                            31,                               // WktStepDuration = 0xFF
}

// ToFIT exports to FIT
// only creates empty FIT object for now
func (w *Workout) ToFIT() (*fit.File, error) {
	h := fit.NewHeader(fit.V10, true)

	workoutmsg := fit.NewWorkoutMsg()
	workoutmsg.WktName = w.Name
	workoutmsg.Sport = sportMapping[w.Sport]

	WorkoutSteps := []*fit.WorkoutStepMsg{}

	for _, step := range w.Steps {
		newmsg := fit.NewWorkoutStepMsg()

		newmsg.MessageIndex = step.MessageIndex
		newmsg.WktStepName = step.WktStepName
		newmsg.DurationType = durationTypes[step.DurationType]
		newmsg.DurationValue = step.DurationValue
		newmsg.Intensity = intensityTypes[step.Intensity]
		newmsg.Notes = step.Notes
		newmsg.TargetType = targetTypes[step.TargetType]
		newmsg.TargetValue = step.TargetValue
		newmsg.CustomTargetValueLow = step.CustomTargetValueLow
		newmsg.CustomTargetValueHigh = step.CustomTargetValueHigh
		WorkoutSteps = append(WorkoutSteps, newmsg)
	}

	workoutFile := fit.WorkoutFile{}
	workoutFile.Workout = workoutmsg
	workoutFile.WorkoutSteps = WorkoutSteps

	newFile, err := fit.NewFile(fit.FileTypeWorkout, h)
	if err != nil {
		return newFile, err
	}

	newFile.FileId.TimeCreated = time.Now()
	newFile.FileId.Manufacturer = fit.ManufacturerGarmin
	newFileWorkoutFile, err := newFile.Workout()
	if err != nil {
		return newFile, err
	}

	*newFileWorkoutFile = workoutFile

	return newFile, nil
}

// MaxUint maximum int value (default in fit)
const MaxUint = ^uint32(0)

func makeStep(s *fit.WorkoutStepMsg) (WorkoutStep, error) {
	step := newWorkoutStep()
	step.MessageIndex = s.MessageIndex
	step.WktStepName = s.WktStepName
	step.DurationType = s.DurationType.String()
	if s.DurationValue < MaxUint {
		step.DurationValue = s.DurationValue
	}
	step.Intensity = s.Intensity.String()
	step.Notes = s.Notes
	step.TargetType = s.TargetType.String()
	if s.TargetValue < MaxUint {
		step.TargetValue = s.TargetValue
	}
	if s.CustomTargetValueLow < MaxUint {
		step.CustomTargetValueLow = s.CustomTargetValueLow
	}
	if s.CustomTargetValueHigh < MaxUint {
		step.CustomTargetValueHigh = s.CustomTargetValueHigh
	}

	return step, nil
}

func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// WriteFit writes FIT file from Workout
func WriteFit(f string, w *fit.File, overwrite bool) (ok bool, err error) {
	if exists(f) && !overwrite {
		err := errors.New("File exists and overwrite was set to false")
		return false, err
	}
	fitFile, err := os.OpenFile(f, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return false, err
	}
	defer fitFile.Close()

	bo := binary.LittleEndian

	err = fit.Encode(fitFile, w, bo)
	if err != nil {
		return false, err
	}

	return true, nil
}

// ReadFit Read FIT file
func ReadFit(f string) (Workout, error) {
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return Workout{}, err
	}

	fitf, err := fit.Decode(bytes.NewReader(data))
	if err != nil {
		return Workout{}, err
	}

	fittype := fitf.FileId.Type

	if fittype != fit.FileTypeWorkout {
		return Workout{}, errors.New("We only accept fit files of type Workout")
	}

	w, err := fitf.Workout()
	if err != nil {
		return Workout{}, err
	}

	steps := w.WorkoutSteps
	if err != nil {
		return Workout{}, err
	}

	neww := Workout{}
	neww.Name = w.Workout.WktName
	// use sportMapping to get the sport name
	for k, v := range sportMapping {
		if v == w.Workout.Sport {
			neww.Sport = k
		}
	}

	neww.Filename = f
	var newsteps []WorkoutStep

	for _, step := range steps {
		s, err := makeStep(step)
		if err != nil {
			return Workout{}, err
		}
		newsteps = append(newsteps, s)
	}

	neww.Steps = newsteps

	return neww, nil
}
