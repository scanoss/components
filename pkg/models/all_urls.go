// SPDX-License-Identifier: GPL-2.0-or-later
/*
 * Copyright (C) 2018-2022 SCANOSS.COM
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 2 of the License, or
 * (at your option) any later version.
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package models

import (
	"context"
	"errors"
	"fmt"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/database"
	purlhelper "github.com/scanoss/go-purl-helper/pkg"
	"go.uber.org/zap"
)

type AllUrlsModel struct {
	ctx context.Context
	s   *zap.SugaredLogger
	q   *database.DBQueryContext
}

type AllUrl struct {
	Component string `db:"component"`
	Version   string `db:"version"`
	License   string `db:"license"`
	LicenseId string `db:"license_id"`
	IsSpdx    bool   `db:"is_spdx"`
	PurlName  string `db:"purl_name"`
	MineId    int32  `db:"mine_id"`
	Url       string `db:"-"`
}

func NewAllUrlModel(ctx context.Context, s *zap.SugaredLogger, q *database.DBQueryContext) *AllUrlsModel {
	return &AllUrlsModel{ctx: ctx, s: s, q: q}
}

func (m *AllUrlsModel) GetUrlsByPurlString(purlString string, limit int) ([]AllUrl, error) {
	if len(purlString) == 0 {
		m.s.Errorf("Please specify a valid Purl String to query")
		return nil, errors.New("please specify a valid Purl String to query")
	}
	purl, err := purlhelper.PurlFromString(purlString)
	if err != nil {
		return nil, err
	}
	purlName, err := purlhelper.PurlNameFromString(purlString) // Make sure we just have the bare minimum for a Purl Name
	if err != nil {
		return nil, err
	}
	return m.GetUrlsByPurlNameType(purlName, purl.Type, limit)
}

func (m *AllUrlsModel) GetUrlsByPurlNameType(purlName, purlType string, limit int) ([]AllUrl, error) {
	if len(purlName) == 0 {
		m.s.Errorf("Please specify a valid Purl Name to query")
		return nil, errors.New("please specify a valid Purl Name to query")
	}
	if len(purlType) == 0 {
		m.s.Errorf("Please specify a valid Purl Type to query: %v", purlName)
		return nil, errors.New("please specify a valid Purl Type to query")
	}

	if limit > defaultMaxVersionLimit || limit <= 0 {
		limit = defaultMaxVersionLimit
	}

	var allUrls []AllUrl
	err := m.q.SelectContext(m.ctx, &allUrls,
		"SELECT component, version,"+
			" l.license_name AS license, l.spdx_id AS license_id, l.is_spdx AS is_spdx,"+
			" purl_name, mine_id FROM all_urls u"+
			" LEFT JOIN mines m ON u.mine_id = m.id"+
			" LEFT JOIN licenses l ON u.license_id = l.id"+
			" WHERE m.purl_type = $1 AND u.purl_name = $2"+
			" ORDER BY date DESC NULLS LAST LIMIT $3",
		purlType, purlName, limit)

	if err != nil {
		m.s.Errorf("Failed to query all urls table for %v - %v: %v", purlType, purlName, err)
		return nil, fmt.Errorf("failed to query the all urls table: %v", err)
	}
	m.s.Debugf("Found %v results for %v, %v.", len(allUrls), purlType, purlName)
	return allUrls, nil
}
