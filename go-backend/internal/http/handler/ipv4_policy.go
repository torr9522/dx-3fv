package handler

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

const ipv4LookupTimeout = 3 * time.Second

type ipv4LookupFunc func(context.Context, string) ([]net.IP, error)

func defaultIPv4Lookup(ctx context.Context, host string) ([]net.IP, error) {
	return net.DefaultResolver.LookupIP(ctx, "ip4", host)
}

func validateIPv4OnlyHost(ctx context.Context, host string) error {
	return validateIPv4OnlyHostWithLookup(ctx, host, defaultIPv4Lookup)
}

func validateIPv4OnlyHostWithLookup(ctx context.Context, host string, lookup ipv4LookupFunc) error {
	host = strings.Trim(strings.TrimSpace(host), "[]")
	if host == "" {
		return errors.New("IPv4 地址或域名不能为空")
	}
	if ip := net.ParseIP(host); ip != nil {
		if ip.To4() == nil {
			return fmt.Errorf("%q 是 IPv6 地址；当前系统默认仅支持 IPv4", host)
		}
		return nil
	}
	if strings.ContainsAny(host, "/?#:") {
		return fmt.Errorf("%q 不是有效的 IPv4 地址或域名", host)
	}
	if lookup == nil {
		return errors.New("IPv4 DNS 解析器不可用")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	lookupCtx, cancel := context.WithTimeout(ctx, ipv4LookupTimeout)
	defer cancel()
	ips, err := lookup(lookupCtx, host)
	if err != nil || len(ips) == 0 {
		return fmt.Errorf("域名 %q 没有可用的 IPv4 A 记录", host)
	}
	for _, ip := range ips {
		if ip != nil && ip.To4() != nil {
			return nil
		}
	}
	return fmt.Errorf("域名 %q 没有可用的 IPv4 A 记录", host)
}

func validateIPv4OnlyNodeAddress(ctx context.Context, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("节点地址不能为空")
	}
	host := value
	if parsed, _, err := net.SplitHostPort(value); err == nil {
		host = parsed
	} else if strings.HasPrefix(value, "[") || strings.Count(value, ":") > 1 {
		host = strings.Trim(value, "[]")
	}
	return validateIPv4OnlyHost(ctx, host)
}

func validateIPv4OnlyListener(value string) error {
	value = strings.Trim(strings.TrimSpace(value), "[]")
	if value == "" {
		return errors.New("监听地址不能为空")
	}
	ip := net.ParseIP(value)
	if ip == nil || ip.To4() == nil {
		return fmt.Errorf("监听地址 %q 必须是 IPv4 地址", value)
	}
	return nil
}

func validateIPv4OnlyRemoteTargets(ctx context.Context, remoteAddr string) error {
	parts := splitRemoteParts(remoteAddr)
	if len(parts) == 0 {
		return errors.New("目标地址不能为空")
	}
	for _, part := range parts {
		host, port, err := parseTargetAddress(part)
		if err != nil || port < 1 || port > 65535 {
			return fmt.Errorf("目标地址 %q 格式无效", part)
		}
		if err := validateIPv4OnlyHost(ctx, host); err != nil {
			return err
		}
	}
	return nil
}

func validateIPv4OnlyHostList(ctx context.Context, value string) error {
	parts := splitRemoteParts(value)
	for _, part := range parts {
		if err := validateIPv4OnlyHost(ctx, part); err != nil {
			return err
		}
	}
	return nil
}

func validateIPv4OnlyTunnelConnectIPs(ctx context.Context, req map[string]interface{}) error {
	return validateIPv4OnlyTunnelConnectIPsWithLegacy(ctx, req, nil)
}

func validateIPv4OnlyTunnelConnectIPsWithLegacy(ctx context.Context, req map[string]interface{}, legacy map[string]struct{}) error {
	groups := [][]map[string]interface{}{asMapSlice(req["outNodeId"])}
	for _, hopRaw := range asAnySlice(req["chainNodes"]) {
		groups = append(groups, asMapSlice(hopRaw))
	}
	for _, group := range groups {
		for _, item := range group {
			connectIP := strings.TrimSpace(asString(item["connectIp"]))
			if connectIP == "" {
				continue
			}
			if _, ok := legacy[connectIP]; ok {
				continue
			}
			if err := validateIPv4OnlyHost(ctx, connectIP); err != nil {
				return fmt.Errorf("自定义连接地址无效: %w", err)
			}
		}
	}
	return nil
}
