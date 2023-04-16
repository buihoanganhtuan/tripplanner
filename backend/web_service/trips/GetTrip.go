package trips

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	cst "github.com/buihoanganhtuan/tripplanner/backend/web_service/_constants"
	utils "github.com/buihoanganhtuan/tripplanner/backend/web_service/_utils"
	"github.com/gorilla/mux"
)

func GetAnonymousTrip(w http.ResponseWriter, rq *http.Request) error {
	id := mux.Vars(rq)["id"]

	if !utils.VerifyBase32String(id, IdLengthChar) {
		return StatusError{
			Status:        InvalidId,
			HttpStatus:    http.StatusBadRequest,
			ClientMessage: InvalidIdMessage,
		}
	}

	ctx := context.Background()
	tx, err := cst.Db.BeginTx(ctx, nil)
	if err != nil {
		return StatusError{
			Status:        DatabaseTransactionError,
			Err:           err,
			HttpStatus:    http.StatusInternalServerError,
			ClientMessage: DatabaseTransactionErrorMessage,
		}
	}

	var budget int
	var name, _expectedDate, _createdDate, _lastModified, budgetUnit, transportMode string
	err = tx.QueryRow("select Name, ExpectedDate, CreatedDate, LastModified, BudgetLimit, BudgetUnit, PreferredTransportMode from ? where id = ?", cst.SqlAnonTripTableVar, id).
		Scan(&name, &_expectedDate, &_createdDate, &_lastModified, &budget, &budgetUnit, &transportMode)
	if err != nil {
		return StatusError{
			Status:        NoSuchTrip,
			Err:           err,
			HttpStatus:    http.StatusBadRequest,
			ClientMessage: NoSuchTripMessage,
		}
	}

	var expectedDate time.Time
	expectedDate, err = time.Parse(cst.DatetimeFormat, _expectedDate)
	if err != nil {
		return StatusError{
			Status:        ParseError,
			Err:           err,
			HttpStatus:    http.StatusInternalServerError,
			ClientMessage: fmt.Sprintf(ParseErrorMessage, "expectedDate"),
		}
	}

	var createdDate time.Time
	createdDate, err = time.Parse(cst.DatetimeFormat, _createdDate)
	if err != nil {
		return StatusError{
			Status:        ParseError,
			Err:           err,
			HttpStatus:    http.StatusInternalServerError,
			ClientMessage: fmt.Sprintf(ParseErrorMessage, "createdDate"),
		}
	}

	var lastModified time.Time
	lastModified, err = time.Parse(cst.DatetimeFormat, _lastModified)
	if err != nil {
		return StatusError{
			Status:        ParseError,
			Err:           err,
			HttpStatus:    http.StatusInternalServerError,
			ClientMessage: fmt.Sprintf(ParseErrorMessage, "lastModified"),
		}
	}

	var edges []Edge
	var rows *sql.Rows
	rows, err = tx.Query("select PointId, NextPointId, Start, DurationHr, DurationMin, CostAmount, CostUnit, Transport from ? order by ord where tripId = ?", cst.SqlEdgeTableVar, id)
	if err != nil {
		return StatusError{
			Status:        DatabaseQueryError,
			Err:           err,
			HttpStatus:    http.StatusInternalServerError,
			ClientMessage: DatabaseQueryErrorMessage,
		}
	}

	for rows.Next() {
		var pointId, nextPointId, _start, costUnit, transport string
		var durationHr, durationMin, costAmount int
		err = rows.Scan(&pointId, &nextPointId, &_start, &durationHr, &durationMin, &costAmount, &costUnit, &transport)
		if err != nil {
			return StatusError{
				Status:        DatabaseQueryError,
				Err:           err,
				HttpStatus:    http.StatusInternalServerError,
				ClientMessage: DatabaseQueryErrorMessage,
			}
		}

		var start time.Time
		start, err = time.Parse(cst.DatetimeFormat, _start)
		if err != nil {
			return StatusError{
				Status:        ParseError,
				Err:           err,
				HttpStatus:    http.StatusInternalServerError,
				ClientMessage: fmt.Sprintf(ParseErrorMessage, "start"),
			}
		}

		edges = append(edges, Edge{
			PointId:     pointId,
			NextPointId: nextPointId,
			Start:       createJsonTime(&start),
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
		return StatusError{
			Status:        DatabaseQueryError,
			Err:           err,
			HttpStatus:    http.StatusInternalServerError,
			ClientMessage: DatabaseQueryErrorMessage,
		}
	}

	var body []byte
	body, err = json.Marshal(Trip{
		Id:           id,
		Type:         "anonymous",
		Name:         name,
		DateExpected: createJsonTime(&expectedDate),
		DateCreated:  createJsonTime(&createdDate),
		LastModified: createJsonTime(&lastModified),
		Budget: Cost{
			Amount: budget,
			Unit:   budgetUnit,
		},
		PreferredMode: transportMode,
		PlanResult:    edges,
	})

	if err != nil {
		return StatusError{
			Status:        UnknownError,
			Err:           err,
			HttpStatus:    http.StatusInternalServerError,
			ClientMessage: UnknownErrorMessage,
		}
	}

	w.Write(body)
	w.WriteHeader(http.StatusOK)
	return nil
}

func GetRegisteredTrip(w http.ResponseWriter, rq *http.Request) error {

}
