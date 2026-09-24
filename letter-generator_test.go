package main

import "testing"

func TestAddressOmitsBlankLines(t *testing.T) {
	tests := []struct {
		name                            string
		nameField, street, postal, city string
		want                            string
	}{
		{
			name:      "complete address",
			nameField: "Max Mustermann",
			street:    "Musterstraße 1",
			postal:    "1010",
			city:      "Wien",
			want:      `Max Mustermann\\Musterstraße 1\\1010 Wien`,
		},
		{
			name:   "missing name and street",
			postal: "1010",
			city:   "Wien",
			want:   "1010 Wien",
		},
		{
			name:      "missing middle fields",
			nameField: "Max Mustermann",
			city:      "Wien",
			want:      `Max Mustermann\\Wien`,
		},
		{
			name: "all fields blank",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := address(tt.nameField, tt.street, tt.postal, tt.city)
			if got != tt.want {
				t.Fatalf("address() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRecipientAddressHeightLines(t *testing.T) {
	tests := []struct {
		name   string
		letter LetterContent
		want   int
	}{
		{
			name:   "one recipient line",
			letter: LetterContent{Recipient: "Stadt Wien"},
			want:   2,
		},
		{
			name: "three recipient lines",
			letter: LetterContent{
				Recipient:           "Stadt Wien",
				RecipientStreet:     "Rathausstraße 1",
				RecipientPostalCode: "1010",
				RecipientCity:       "Wien",
			},
			want: 4,
		},
		{
			name: "empty recipient still reserves one line",
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.letter.RecipientAddressHeightLines(); got != tt.want {
				t.Fatalf("RecipientAddressHeightLines() = %d, want %d", got, tt.want)
			}
		})
	}
}
