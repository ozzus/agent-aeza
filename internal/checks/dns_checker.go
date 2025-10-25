package checks

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"ozzus/agent-aeza/internal/domain"
)

type DNSChecker struct {
	baseMetadata
	timeout time.Duration
}

func NewDNSChecker(timeout time.Duration, location, country string) *DNSChecker {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	return &DNSChecker{
		baseMetadata: newBaseMetadata(location, country),
		timeout:      timeout,
	}
}

func (d *DNSChecker) Check(target string, parameters map[string]interface{}) (*domain.CheckResult, error) {
	if target == "" {
		return &domain.CheckResult{Status: domain.StatusFailed, Error: "empty target"}, nil
	}

	recordType := strings.ToUpper(stringParam(parameters, "record_type", string(domain.DNSRecordA)))
	timeout := durationParam(parameters, "timeout", d.timeout)
	if timeout <= 0 {
		timeout = d.timeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resolver := net.Resolver{}

	records, err := d.lookupRecords(ctx, resolver, target, recordType)
	if err != nil {
		return &domain.CheckResult{Status: domain.StatusFailed, Error: err.Error()}, nil
	}

	payload := map[string]interface{}{
		"dns": map[string]interface{}{
			"locations": []map[string]interface{}{
				{
					"location": d.locationValue(parameters),
					"country":  d.countryValue(parameters),
					"records":  strings.Join(records, ", "),
					"ttl":      "N/A",
				},
			},
		},
	}

	return &domain.CheckResult{
		Status:  domain.StatusSuccess,
		Payload: payload,
	}, nil
}

func (d *DNSChecker) lookupRecords(ctx context.Context, resolver net.Resolver, target, recordType string) ([]string, error) {
	switch domain.DNSRecordType(recordType) {
	case domain.DNSRecordA:
		addrs, err := resolver.LookupIPAddr(ctx, target)
		if err != nil {
			return nil, err
		}
		return d.formatIPAddrs(addrs, false), nil
	case domain.DNSRecordAAAA:
		addrs, err := resolver.LookupIPAddr(ctx, target)
		if err != nil {
			return nil, err
		}
		return d.formatIPAddrs(addrs, true), nil
	case domain.DNSRecordMX:
		mx, err := resolver.LookupMX(ctx, target)
		if err != nil {
			return nil, err
		}
		sort.Slice(mx, func(i, j int) bool { return mx[i].Pref < mx[j].Pref })
		results := make([]string, 0, len(mx))
		for _, entry := range mx {
			results = append(results, fmt.Sprintf("%d %s", entry.Pref, strings.TrimSuffix(entry.Host, ".")))
		}
		return results, nil
	case domain.DNSRecordNS:
		ns, err := resolver.LookupNS(ctx, target)
		if err != nil {
			return nil, err
		}
		results := make([]string, 0, len(ns))
		for _, entry := range ns {
			results = append(results, strings.TrimSuffix(entry.Host, "."))
		}
		return results, nil
	case domain.DNSRecordTXT:
		txt, err := resolver.LookupTXT(ctx, target)
		if err != nil {
			return nil, err
		}
		return txt, nil
	default:
		return nil, fmt.Errorf("unsupported record type: %s", recordType)
	}
}

func (d *DNSChecker) formatIPAddrs(addrs []net.IPAddr, ipv6 bool) []string {
	results := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		ip := addr.IP
		if ipv6 && ip.To4() != nil {
			continue
		}
		if !ipv6 && ip.To4() == nil {
			continue
		}
		results = append(results, ip.String())
	}
	if len(results) == 0 {
		for _, addr := range addrs {
			results = append(results, addr.IP.String())
		}
	}
	return results
}

func (d *DNSChecker) Type() domain.TaskType {
	return domain.TaskTypeDNS
}
