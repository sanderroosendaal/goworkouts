package goworkouts

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"time"

	"github.com/tormoder/fit"
)

// WorkoutStep is the container of a Workout Step
type workoutStep struct {
	MessageIndex          fit.MessageIndex `json:"stepId"`
	WktStepName           string           `json:"wkt_step_name"`
	DurationType          string           `json:"durationType"`
	DurationValue         uint32           `json:"durationValue"`
	TargetType            string           `json:"targetType"`
	TargetValue           uint32           `json:"targetValue"`
	CustomTargetValueLow  uint32           `json:"targetValueLow"`
	CustomTargetValueHigh uint32           `json:"targetValueHigh"`
	Intensity             string           `json:"intensity"`
	Notes                 string           `json:"description"`
	// Type                  string           `json:"type"`
}

// NewWorkoutStep creates new workout step
func newWorkoutStep() workoutStep {
	newstep := workoutStep{}
	newstep.DurationValue = 0
	newstep.TargetValue = 0
	newstep.CustomTargetValueLow = 0
	newstep.CustomTargetValueHigh = 0
	return newstep
}

// Workout is a Workout
type Workout struct {
	Filename    string        `json:"filename"`
	Name        string        `json:"workoutName"`
	Steps       []workoutStep `json:"steps"`
	Sport       string        `json:"sport"`
	Description string        `json:"description"`
	// WorkoutID uint64         `json:"WorkoutId"`
	// OwnerID   uint64         `json:"ownerId"`
}

// ToJSON export to JSON
func (w *Workout) ToJSON() ([]byte, error) {
	return json.Marshal(w)
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

	workoutsteps := []*fit.WorkoutStepMsg{}

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
		workoutsteps = append(workoutsteps, newmsg)
	}

	workoutFile := fit.WorkoutFile{}
	workoutFile.Workout = workoutmsg
	workoutFile.WorkoutSteps = workoutsteps

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

func makeStep(s *fit.WorkoutStepMsg) (workoutStep, error) {
	step := newWorkoutStep()
	step.MessageIndex = s.MessageIndex
	step.WktStepName = s.WktStepName
	step.DurationType = s.DurationType.String()
	step.DurationValue = s.DurationValue
	step.Intensity = s.Intensity.String()
	step.Notes = s.Notes
	step.TargetType = s.TargetType.String()
	step.TargetValue = s.TargetValue
	step.CustomTargetValueLow = s.CustomTargetValueLow
	step.CustomTargetValueHigh = s.CustomTargetValueHigh

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
	neww.Filename = f
	var newsteps []workoutStep

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
