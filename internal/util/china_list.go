package util

import (
	"bufio"
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/rs/zerolog"
)

type ChinaList struct {
	ctx        context.Context
	httpClient *http.Client
}

const CHINA_LIST_URL = "https://raw.githubusercontent.com/felixonmars/dnsmasq-china-list/master/accelerated-domains.china.conf"

func MakeChinaList(ctx context.Context, proxy string) *ChinaList {
	chinaList := ChinaList{
		ctx:        ctx,
		httpClient: new(http.Client),
	}

	if proxy != "" {
		proxyUrl, err := url.Parse(proxy)
		if err != nil {
			panic(err)
		}
		chinaList.httpClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
	}

	return &chinaList
}

func (c *ChinaList) Fetch() ([]string, error) {
	logger := zerolog.Ctx(c.ctx).
		With().
		Str("module", "china_list").
		Logger()

	req, err := http.NewRequestWithContext(c.ctx, "GET", CHINA_LIST_URL, http.NoBody)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to create request")
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to send request")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logger.Error().Stack().Err(err).Int("StatusCode", resp.StatusCode).Msg("StatusCode")
		return nil, err
	}

	chinalist := make([]string, 0, 64500)

	buf := bufio.NewScanner(resp.Body)
	for buf.Scan() {
		line := buf.Text()
		if line[0] == '#' {
			continue
		}
		e := strings.LastIndexByte(line, '/')
		chinalist = append(chinalist, line[8:e])
	}

	return chinalist, nil
}
