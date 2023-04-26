package trips

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	cst "github.com/buihoanganhtuan/tripplanner/backend/web_service/_constants"
	"github.com/gorilla/mux"
)

func NewGetTripHandler(anonymous bool) cst.ErrorHandler {
	var eh cst.ErrorHandler
	eh = cst.ErrorHandler(func(w http.ResponseWriter, r *http.Request) (error, *cst.ErrorResponse) {
		id := mux.Vars(r)["id"]

		ctx := context.Background()
		tx, err := cst.Db.BeginTx(ctx, nil)
		if err != nil {
			return err, newDatabaseQueryError()
		}

		var budget int
		var uid, name, edStr, cdStr, lmStr, budgetUnit, tm string
		if anonymous {

		}
		if anonymous {
			r := tx.QueryRow("select Name, LastModified, BudgetLimit, BudgetUnit, PreferredTransportMode from ? where id = ?", cst.Ev.Var(cst.SqlAnonTripTableVar), id)
			err = r.Scan(&name, &lmStr, &budget, &budgetUnit, &tm)
		} else {
			r := tx.QueryRow("select UserId, Name, ExpectedDate, CreatedDate, LastModified, BudgetLimit, BudgetUnit, PreferredTransportMode from ? where id = ?", cst.Ev.Var(cst.SqlTripTableVar), id)
			err = r.Scan(&uid, &name, &edStr, &cdStr, &lmStr, &budget, &budgetUnit, &tm)
		}

		if err != nil {
			return err, newDatabaseQueryError()
		}

		var ed, cd, lastModified time.Time
		var expectedDate, createdDate *time.Time

		if !anonymous {
			ed, err = time.Parse(cst.DatetimeFormat, edStr)
			if err != nil {
				return err, newServerParseError()
			}
			expectedDate = &ed

			cd, err = time.Parse(cst.DatetimeFormat, cdStr)
			if err != nil {
				return err, newServerParseError()
			}
			createdDate = &cd
		}

		lastModified, err = time.Parse(cst.DatetimeFormat, lmStr)
		if err != nil {
			return err, newServerParseError()
		}

		var edges []Edge
		var rows *sql.Rows
		rows, err = tx.Query("select PointId, NextPointId, Start, DurationHr, DurationMin, CostAmount, CostUnit, Transport from ? order by ord where tripId = ?", cst.SqlEdgeTableVar, id)
		if err != nil {
			return err, newDatabaseQueryError()
		}

		for rows.Next() {
			var pointId, nextPointId, _start, costUnit, transport string
			var durationHr, durationMin, costAmount int
			err = rows.Scan(&pointId, &nextPointId, &_start, &durationHr, &durationMin, &costAmount, &costUnit, &transport)
			if err != nil {
				return err, newDatabaseQueryError()
			}

			var start time.Time
			start, err = time.Parse(cst.DatetimeFormat, _start)
			if err != nil {
				return err, newServerParseError()
			}

			edges = append(edges, Edge{
				PointId:     PointId(pointId),
				NextPointId: PointId(nextPointId),
				Start:       cst.JsonDateTime(start),
				Duration: Duration{
					Hour: durationHr,
					Min:  durationMin,
				},
				Cost: Cost{
					Amount: costAmount,
					Unit:   costUnit,
				},
				Transport: transport,
			})
		}

		err = rows.Err()
		if err != nil {
			return err, newDatabaseQueryError()
		}

		var body []byte
		tp := "registered"
		if anonymous {
			tp = "anonymous"
		}
		body, err = json.Marshal(Trip{
			Id:           id,
			UserId:       uid,
			Type:         tp,
			Name:         name,
			DateExpected: (*cst.JsonDateTime)(expectedDate),
			DateCreated:  (*cst.JsonDateTime)(createdDate),
			LastModified: cst.JsonDateTime(lastModified),
			Budget: Cost{
				Amount: budget,
				Unit:   budgetUnit,
			},
			PreferredMode: tm,
			PlanResult:    edges,
		})

		if err != nil {
			return err, newUnknownError()
		}

		w.Write(body)
		w.WriteHeader(http.StatusOK)
		return nil, nil
	})
	return eh
}
