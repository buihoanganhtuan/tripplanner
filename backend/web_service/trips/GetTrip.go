package trips

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	cst "github.com/buihoanganhtuan/tripplanner/backend/web_service/_constants"
	utils "github.com/buihoanganhtuan/tripplanner/backend/web_service/_utils"
	"github.com/gorilla/mux"
)

func NewGetTripHandler(anonymous bool) utils.ErrorHandler {
	var eh utils.ErrorHandler
	eh = utils.ErrorHandler(func(w http.ResponseWriter, r *http.Request) (error, utils.ErrorResponse) {
		id := mux.Vars(r)["id"]

		ctx := context.Background()
		tx, err := cst.Db.BeginTx(ctx, nil)
		if err != nil {
			return err, utils.NewDatabaseQueryError()
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
			return err, utils.NewDatabaseQueryError()
		}

		var ed, cd, lastModified time.Time
		var expectedDate, createdDate *time.Time

		if !anonymous {
			ed, err = time.Parse(cst.DatetimeFormat, edStr)
			if err != nil {
				return err, utils.NewServerParseError()
			}
			expectedDate = &ed

			cd, err = time.Parse(cst.DatetimeFormat, cdStr)
			if err != nil {
				return err, utils.NewServerParseError()
			}
			createdDate = &cd
		}

		lastModified, err = time.Parse(cst.DatetimeFormat, lmStr)
		if err != nil {
			return err, utils.NewServerParseError()
		}

		var edges []Edge
		var rows *sql.Rows
		rows, err = tx.Query("select PointId, NextPointId, Start, DurationHr, DurationMin, CostAmount, CostUnit, Transport from ? order by ord where tripId = ?", cst.SqlEdgeTableVar, id)
		if err != nil {
			return err, utils.NewDatabaseQueryError()
		}

		for rows.Next() {
			var pointId, nextPointId, _start, costUnit, transport string
			var durationHr, durationMin, costAmount int
			err = rows.Scan(&pointId, &nextPointId, &_start, &durationHr, &durationMin, &costAmount, &costUnit, &transport)
			if err != nil {
				return err, utils.NewDatabaseQueryError()
			}

			var start time.Time
			start, err = time.Parse(cst.DatetimeFormat, _start)
			if err != nil {
				return err, utils.NewServerParseError()
			}

			edges = append(edges, Edge{
				PointId:     PointId(pointId),
				NextPointId: PointId(nextPointId),
				Start:       utils.JsonDateTime(start),
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
			return err, utils.NewDatabaseQueryError()
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
			DateExpected: (*utils.JsonDateTime)(expectedDate),
			DateCreated:  (*utils.JsonDateTime)(createdDate),
			LastModified: utils.JsonDateTime(lastModified),
			Budget: Cost{
				Amount: budget,
				Unit:   budgetUnit,
			},
			PreferredMode: tm,
			PlanResult:    edges,
		})

		if err != nil {
			return err, utils.NewUnknownError()
		}

		w.Write(body)
		w.WriteHeader(http.StatusOK)
		return nil, utils.ErrorResponse{}
	})
	return eh
}
