package goworkouts

import (
	"bytes"
	"encoding/json"
	"errors"

	"os"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/muktihari/fit/encoder"
	"github.com/muktihari/fit/decoder"
	//"github.com/muktihari/fit/profile"
	//"github.com/muktihari/fit/profile/untyped/fieldnum"
	"github.com/muktihari/fit/profile/typedef"
	"github.com/muktihari/fit/profile/filedef"
	"github.com/muktihari/fit/profile/mesgdef"
	"github.com/muktihari/fit/proto"

	yaml "gopkg.in/yaml.v2"
)

// Mapping of string-based sport names to fit.Sport values

var sportMapping = map[string]typedef.Sport{
	"generic":   typedef.SportGeneric,
	"running":   typedef.SportRunning,
	"cycling":   typedef.SportCycling,
	"swimming":  typedef.SportSwimming,
	"walking":   typedef.SportWalking,
	"crosscountryski": typedef.SportCrossCountrySkiing,
	"rowing":    typedef.SportRowing,
	"hiking":    typedef.SportHiking,
	"multi":     typedef.SportMultisport,
	"inlineskate": typedef.SportInlineSkating,
	"iceskate": typedef.SportIceSkating,
	"hitt": typedef.SportHiit,
}


// WorkoutStep is the container of a Workout Step
type WorkoutStep struct {
	MessageIndex          typedef.MessageIndex `json:"stepId" yaml:"stepId"`
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

func GetStepByIndex(w Workout, idx typedef.MessageIndex) (WorkoutStep, error) {
	for _, step := range w.Steps {
		if step.MessageIndex == idx {
			return step, nil
		}
	}
	return WorkoutStep{}, errors.New("Step not found")
}

func AddRepeats(stepslist []string, idxlist []typedef.MessageIndex, idx typedef.MessageIndex, nr_repeats uint32) ([]string) {
	for i, _ := range(stepslist) {
		if idxlist[i] == idx {
			stepslist[i] = fmt.Sprintf("\n%vx\n", nr_repeats)+stepslist[i]
		}
	}
	return stepslist
}

// Helper function for ToIntervals
func TransformRepeats(input string) string {
	lines := strings.Split(input, "\n")
	re := regexp.MustCompile(`^(\d+)x$`)
	i := 0
	var result []string
	
	for i < len(lines) {
		stopped := false
		multiplier := 1
		currentLine := strings.TrimSpace(lines[i])
		// Preserve empty lines but skip them for merging logic
		if currentLine == "" {
			result = append(result, "")
			i++
			continue
		}
		currentMatch := re.FindStringSubmatch(currentLine)
		if currentMatch == nil {
			result = append(result, strings.TrimSpace(lines[i]))
			i++
			continue
		}
		// Still here, we have a match. Need to update multiplier and scan next line(s)
		multiplier, _ = strconv.Atoi(currentMatch[1])
		for {
			if i+1 < len(lines) {
				// loop until non-empty line found
				nextLine := ""
				for {
					nextLine = strings.TrimSpace(lines[i+1])
					if nextLine != "" {
						break
					}
					i++
					if i+1 >= len(lines) {
						break
					}
				}
				// check if it matches a repeat step
				nextMatch := re.FindStringSubmatch(nextLine)
				if nextMatch == nil {
					result = append(result, fmt.Sprintf("\n%dx", multiplier))
					result = append(result, strings.TrimSpace(nextLine))
					i += 2
					multiplier = 1
					stopped = true
				}
				if nextMatch != nil { // found a multiplier
					n, _ := strconv.Atoi(nextMatch[1])
					multiplier *= n
					i++
				}
				if stopped {
					multiplier = 1
					break
				}
			}
		}
		//i++
	}
			
	return strings.Join(result, "\n")
}

// ToIntervals exports to intervals.icu workout description language
// Each step is turned into a string like "- 10m @ 200W Comment"
func (w *Workout) ToIntervals() (string, error) {
	var err error
	var stepstext string
	var stepslist []string
	var idxlist []typedef.MessageIndex
	for _, step := range w.Steps {
		idxlist = append(idxlist, step.MessageIndex)
		var buffer bytes.Buffer
		var prefix, duration, target, name, notes, intensity string
		if step.DurationType == "RepeatUntilStepsCmplt" {
			nr_repeats := step.TargetValue
			idx := typedef.MessageIndex(step.DurationValue)
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
				target = "ramp Z1-Z2"
			}
			if step.Intensity == "Cooldown" {
				prefix = "\nCooldown\n"
				target = "ramp Z2-Z1"
			}

			if step.Intensity == "Recovery" {
				target = "Z1"
			}

			if step.Intensity == "Rest" {
				target = "Z1"
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
	stepstext = TransformRepeats(stepstext)
	
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

var targetTypes = map[string]typedef.WktStepTarget{
	"Speed":        typedef.WktStepTargetSpeed,        //        WktStepTarget = 0
	"HeartRate":    typedef.WktStepTargetHeartRate,    //    WktStepTarget = 1
	"Open":         typedef.WktStepTargetOpen,         //         WktStepTarget = 2
	"Cadence":      typedef.WktStepTargetCadence,      //      WktStepTarget = 3
	"Power":        typedef.WktStepTargetPower,        //        WktStepTarget = 4
	"Grade":        typedef.WktStepTargetGrade,        //        WktStepTarget = 5
	"Resistance":   typedef.WktStepTargetResistance,   //   WktStepTarget = 6
	"Power3s":      typedef.WktStepTargetPower3S,      //   WktStepTarget = 7
	"Power10s":     typedef.WktStepTargetPower10S,     //    WktStepTarget = 8
	"Power30s":     typedef.WktStepTargetPower30S,     //   WktStepTarget = 9
	"PowerLap":     typedef.WktStepTargetPowerLap,     //   WktStepTarget = 10
	"SwimStroke":   typedef.WktStepTargetSwimStroke,   //  WktStepTarget = 11
	"SpeedLap":     typedef.WktStepTargetSpeedLap,     // WktStepTarget = 12
	"HeartRateLap": typedef.WktStepTargetHeartRateLap, // WktStepTarget = 13
}

var intensityTypes = map[string]typedef.Intensity{
	"Active":   typedef.IntensityActive,   //   Intensity = 0
	"Rest":     typedef.IntensityRest,     //  Intensity = 1
	"Warmup":   typedef.IntensityWarmup,   // Intensity = 2
	"Cooldown": typedef.IntensityCooldown, // Intensity = 3
	"Recovery": typedef.IntensityRecovery, // Intensity = 4
	"Interval": typedef.IntensityInterval, // Intensity = 5
	"Other": typedef.IntensityOther, // Intensity = 6
	//IntensityInvalid  Intensity = 0xFF
}

var durationTypes = map[string]typedef.WktStepDuration{
	"Time":                               typedef.WktStepDurationTime,
	"Distance":                           typedef.WktStepDurationDistance,
	"HrLessThan":                         typedef.WktStepDurationHrLessThan,                      //                         // WktStepDuration = 2
	"HrGreaterThan":                      typedef.WktStepDurationHrGreaterThan,                   //                      // WktStepDuration = 3
	"Calories":                           typedef.WktStepDurationCalories,                        //                           // WktStepDuration = 4
	"Open":                               typedef.WktStepDurationOpen,                            //                               // WktStepDuration = 5
	"RepeatUntilStepsCmplt":              typedef.WktStepDurationRepeatUntilStepsCmplt,           // WktStepDuration = 6
	"RepeatUntilTime":                    typedef.WktStepDurationRepeatUntilTime,                 // WktStepDuration = 7
	"RepeatUntilDistance":                typedef.WktStepDurationRepeatUntilDistance,             // WktStepDuration = 8
	"RepeatUntilCalories":                typedef.WktStepDurationRepeatUntilHrLessThan,           // WktStepDuration = 9
	"RepeatUntilHrLessThan":              typedef.WktStepDurationRepeatUntilHrLessThan,           // WktStepDuration = 10
	"RepeatUntilHrGreaterThan":           typedef.WktStepDurationRepeatUntilHrGreaterThan,        // WktStepDuration = 11
	"RepeatUntilPowerLessThan":           typedef.WktStepDurationRepeatUntilPowerLessThan,        // WktStepDuration = 12
	"RepeatUntilPowerGreaterThan":        typedef.WktStepDurationRepeatUntilPowerGreaterThan,     // WktStepDuration = 13
	"PowerLessThan":                      typedef.WktStepDurationPowerLessThan,                   // WktStepDuration = 14
	"PowerGreaterThan":                   typedef.WktStepDurationPowerGreaterThan,                // WktStepDuration = 15
	"TrainingPeaksTss":                   typedef.WktStepDurationTrainingPeaksTss,                // WktStepDuration = 16
	"RepeatUntilPowerLastLapLessThan":    typedef.WktStepDurationRepeatUntilPowerLastLapLessThan, // WktStepDuration = 17
	"RepeatUntilMaxPowerLastLapLessThan": typedef.WktStepDurationRepeatUntilPowerLastLapLessThan, // WktStepDuration = 18
	"Power3sLessThan":                    typedef.WktStepDurationPower3SLessThan,                 // WktStepDuration = 19
	"Power10sLessThan":                   typedef.WktStepDurationPower10SLessThan,                // WktStepDuration = 20
	"Power30sLessThan":                   typedef.WktStepDurationPower30SLessThan,                // WktStepDuration = 21
	"Power3sGreaterThan":                 typedef.WktStepDurationPower3SGreaterThan,              // WktStepDuration = 22
	"Power10sGreaterThan":                typedef.WktStepDurationPower10SGreaterThan,             // WktStepDuration = 23
	"Power30sGreaterThan":                typedef.WktStepDurationPower30SGreaterThan,             // WktStepDuration = 24
	"PowerLapLessThan":                   typedef.WktStepDurationPowerLapLessThan,                // WktStepDuration = 25
	"PowerLapGreaterThan":                typedef.WktStepDurationPowerLapGreaterThan,             // WktStepDuration = 26
	"RepeatUntilTrainingPeaksTss":        typedef.WktStepDurationRepeatUntilTrainingPeaksTss,     // WktStepDuration = 27
	"RepetitionTime":                     typedef.WktStepDurationRepetitionTime,                  // WktStepDuration = 28
	"Reps":                               typedef.WktStepDurationReps,                            // WktStepDuration = 29
	"TimeOnly":                           typedef.WktStepDurationTimeOnly,                        // WktStepDuration = 31
	//"Invalid":                            31,                               // WktStepDuration = 0xFF
}

// ToFIT exports to FIT
func (w *Workout) ToFIT() (proto.FIT, error) {


	WorkoutSteps := []*mesgdef.WorkoutStep{}

	for _, step := range w.Steps {
		newstep := mesgdef.NewWorkoutStep(nil)

		newstep.MessageIndex = step.MessageIndex
		newstep.WktStepName = step.WktStepName
		newstep.DurationType = durationTypes[step.DurationType]
		newstep.DurationValue = step.DurationValue
		newstep.Intensity = intensityTypes[step.Intensity]
		newstep.Notes = step.Notes
		newstep.TargetType = targetTypes[step.TargetType]
		newstep.TargetValue = step.TargetValue
		newstep.CustomTargetValueLow = step.CustomTargetValueLow
		newstep.CustomTargetValueHigh = step.CustomTargetValueHigh
		WorkoutSteps = append(WorkoutSteps, newstep)
	}

	workout := filedef.NewWorkout()
	workout.FileId = *mesgdef.NewFileId(nil).
		SetType(typedef.FileWorkout)
	
	fitw := mesgdef.NewWorkout(nil)
	fitw.SetWktName(w.Name)
	fitw.SetSport(sportMapping[w.Sport])
	workoutmsg := fitw.ToMesg(nil)
	
	workout.Add(workoutmsg)
	for _, wktstep := range WorkoutSteps {
		workout.Add(wktstep.ToMesg(nil))
	}


	return workout.ToFIT(nil), nil
}

// MaxUint maximum int value (default in fit)
const MaxUint = ^uint32(0)

func makeStep(s *mesgdef.WorkoutStep) (WorkoutStep, error) {
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
func WriteFit(f string, w *proto.FIT, overwrite bool) (ok bool, err error) {
	if exists(f) && !overwrite {
		err := errors.New("File exists and overwrite was set to false")
		return false, err
	}
	fitFile, err := os.OpenFile(f, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return false, err
	}
	defer fitFile.Close()

	enc := encoder.New(fitFile)
	if err := enc.Encode(&w); err != nil {
		return false, err
	}

	return true, nil
}

// ReadFit Read FIT file
func ReadFit(f string) (Workout, error) {
	fitFile, err := os.Open(f)
	if err != nil {
		return Workout{}, err
	}
	defer fitFile.Close()
	
	lis := filedef.NewListener()
	defer lis.Close()
	
	dec := decoder.New(fitFile,
		decoder.WithMesgListener(lis),
		decoder.WithBroadcastOnly(),
	)
	_, err = dec.Decode()
	if err != nil {
		return Workout{}, err
	}

	fitf, ok := lis.File().(*filedef.Workout)
	
	if !ok {
		return Workout{}, errors.New("We only accept fit files of type Workout")
	}

	w := fitf.Workout

	steps := fitf.WorkoutSteps

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
