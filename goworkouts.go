package goworkouts

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/tormoder/fit"
)

// WorkoutStep is the container of a Workout Step
type WorkoutStep struct {
	MessageIndex          fit.MessageIndex `json:"name=message_index"`
	WktStepName           string           `json:"wkt_step_name"`
	DurationType          string           `json:"duration_type"`
	DurationValue         uint32           `json:"duration_value"`
	TargetType            string           `json:"target_type"`
	TargetValue           uint32           `json:"target_value"`
	CustomTargetValueLow  uint32           `json:"custom_target_value_low"`
	CustomTargetValueHigh uint32           `json:"custom_target_value_high"`
	Intensity             string           `json:"intensity"`
	Notes                 string           `json:"notes"`
}

// Workout is a Workout
type Workout struct {
	Filename string        `json:"filename"`
	Name     string        `json:"name"`
	Steps    []WorkoutStep `json:"steps"`
}

// ToJSON export to JSON
func (w *Workout) ToJSON() ([]byte, error) {
	return json.Marshal(w)
}

func makeStep(s *fit.WorkoutStepMsg) (WorkoutStep, error) {
	step := WorkoutStep{}
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
