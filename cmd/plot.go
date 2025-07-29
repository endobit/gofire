package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"endobit.io/gofire"
)

type Event struct {
	Event string    `json:"event"`
	Time  time.Time `json:"time"`
}

func newPlotCmd() *cobra.Command {
	var (
		input  string
		output string
		events string
	)

	cmd := cobra.Command{
		Use:   "plot",
		Short: "Create a scatter plot from a previous run",
		RunE: func(_ *cobra.Command, _ []string) error {
			fin, err := os.Open(input)
			if err != nil {
				return fmt.Errorf("failed to open input file %q: %w", input, err)
			}
			defer fin.Close()

			var temps []gofire.Status

			for s := bufio.NewScanner(fin); s.Scan(); { // log isn't json, but each line is
				var status gofire.Status

				if err := json.Unmarshal(s.Bytes(), &status); err != nil {
					return err
				}

				temps = append(temps, status)
			}

			var markers []gofire.Marker

			if events != "" {
				fin, err := os.Open(events)
				if err != nil {
					return fmt.Errorf("failed to open events file %q: %w", events, err)
				}
				defer fin.Close()

				var events []Event

				if err := json.NewDecoder(fin).Decode(&events); err != nil {
					return fmt.Errorf("failed to decode events file %q: %w", events, err)
				}

				for _, e := range events {
					markers = append(markers, gofire.Marker{
						Time:  e.Time,
						Label: e.Event,
					})
				}
			}

			p := gofire.NewPlotter(&gofire.PlotterOptions{
				Title:   temps[0].Time.Format(time.ANSIC),
				Data:    temps,
				Markers: markers,
			})

			plot, err := p.Plot()
			if err != nil {
				return err
			}

			return plot.Save(800, 300, output)
		},
	}

	cmd.Flags().StringVarP(&input, "input", "i", "", "input file")
	cmd.Flags().StringVarP(&output, "output", "o", "gofire.png", "output file")
	cmd.Flags().StringVar(&events, "events", "", "events file")

	if err := cmd.MarkFlagRequired("input"); err != nil {
		panic(err)
	}

	return &cmd
}
