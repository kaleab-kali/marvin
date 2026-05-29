package cost

import "testing"

func TestFilterIgnoredServices(t *testing.T) {
	records := []Record{
		{Service: "Amazon EC2", Cost: 100},
		{Service: "Tax", Cost: 10},
		{Service: "Credits", Cost: -5},
	}

	got := FilterIgnoredServices(records, []string{"Tax", "Credits"})

	if len(got) != 1 {
		t.Fatalf("expected 1 record, got %d", len(got))
	}
	if got[0].Service != "Amazon EC2" {
		t.Fatalf("expected Amazon EC2, got %q", got[0].Service)
	}
}

func TestFilterIgnoredServicesReturnsOriginalRecordsWhenNoRules(t *testing.T) {
	records := []Record{{Service: "Amazon EC2", Cost: 100}}

	got := FilterIgnoredServices(records, nil)

	if len(got) != 1 {
		t.Fatalf("expected 1 record, got %d", len(got))
	}
	if got[0].Service != "Amazon EC2" {
		t.Fatalf("expected Amazon EC2, got %q", got[0].Service)
	}
}

func TestFilterIncludedServices(t *testing.T) {
	records := []Record{
		{Service: "Amazon EC2", Cost: 100},
		{Service: "Amazon S3", Cost: 25},
		{Service: "Tax", Cost: 10},
	}

	got := FilterIncludedServices(records, []string{"Amazon EC2", "Amazon S3"})

	if len(got) != 2 {
		t.Fatalf("expected 2 records, got %d", len(got))
	}
	if got[0].Service != "Amazon EC2" || got[1].Service != "Amazon S3" {
		t.Fatalf("unexpected included services: %+v", got)
	}
}

func TestFilterIncludedServicesReturnsOriginalRecordsWhenNoRules(t *testing.T) {
	records := []Record{{Service: "Amazon EC2", Cost: 100}}

	got := FilterIncludedServices(records, nil)

	if len(got) != 1 {
		t.Fatalf("expected 1 record, got %d", len(got))
	}
	if got[0].Service != "Amazon EC2" {
		t.Fatalf("expected Amazon EC2, got %q", got[0].Service)
	}
}
