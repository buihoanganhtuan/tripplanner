import React, { useState } from "react"
import { MapBox } from "./MapBox"
import { PlanningBox, Point } from "./PlanningBox"
import { SearchBox } from "./SearchBox"
import { StagingBox } from "./StagingBox"

interface PlanningPaneState {
    stagingPoints: GeoPoint[]
    tripPoints: Point[]
    selectedStagingPoint: GeoPoint | null
    selectedTripPoint: Point | null
    autocompletedPoint: GeoPoint | null
}

interface Address {
    prefecture: string
    city: string
    district?: string
    landcode?: string
}

export interface GeoPoint {
    id: string
    name: string
    address: Address
    lat: number
    lon: number
}

export function PlanningPane() {
    const [state, setState] = useState<PlanningPaneState>({
        stagingPoints: [],
        tripPoints: [],
        selectedStagingPoint: null,
        selectedTripPoint: null,
        autocompletedPoint: null
    })

    const handleStagingPointSelection = (point: GeoPoint) => {
        if (state.selectedStagingPoint == point)
            return
        setState(prev => { return { ...prev, selectedStagingPoint: point, selectedTripPoint: null } })
    }

    const handleTripPointSelection = (point: Point) => {
        if (state.selectedTripPoint == point)
            return
        setState(prev => { return { ...prev, selectedTripPoint: point, selectedStagingPoint: null} })
    }

    const handleAutocompletion = (point: GeoPoint) => {
        for (let p of state.stagingPoints) {
            if (p.id == point.id)
                return
        }
        setState(prev => {
            prev.stagingPoints.push(point)
            return { ...prev }
        })
    }

    const handleStagingPointDeletion = (point: GeoPoint) => {
        setState(prev => {
            return {
                ...prev,
                stagingPoints: prev.stagingPoints.filter(p => p != point),
                selectedStagingPoint: point == prev.selectedStagingPoint ? null : prev.selectedStagingPoint
            }
        })
    }

    const handleStagingPointTransfer = (point: GeoPoint) => {
        if (state.tripPoints.length > 0 && state.tripPoints.map(p => p.id === point.id).reduce((prev, cur) => prev || cur)) {
            alert("Point already in trip")
            return
        }

        let nextSelectedStagingPoint = state.selectedStagingPoint != null && state.selectedStagingPoint == point ? null : state.selectedStagingPoint
        let nextStagingPoints = state.stagingPoints.filter(p => p != point)
        let nextTripPoints = state.tripPoints
        nextTripPoints.push({
            ...point,
            before: [],
            after: [],
            isFirst: false,
            isLast: false,
            arriveBefore: "",
            stayDuration: 0,
        })
        setState(prev => {
            return {
                ...prev,
                stagingPoints: nextStagingPoints,
                selectedStagingPoint: nextSelectedStagingPoint,
                tripPoints: nextTripPoints
            }
        })
    }

    const handleTripPointDeletion = (point: Point) => {
        setState(prev => {
            return {
                ...prev,
                tripPoints: prev.tripPoints.filter(p => p != point),
                selectedTripPoint: point == prev.selectedTripPoint ? null : prev.selectedTripPoint
            }
        })
    }

    return (
        <div className="grid grid-row-4 justify-items-center items-center gap-y-4">
            <SearchBox input="" selectedEntry={state.autocompletedPoint} onEntrySelection={handleAutocompletion} className="row-start-1"/>
            <StagingBox points={state.stagingPoints} 
                        selectedPoint={state.selectedStagingPoint}
                        onPointDelete={handleStagingPointDeletion}
                        onPointSelect={handleStagingPointSelection}
                        onPointTransfer={handleStagingPointTransfer}
                        className="row-start-2"/>
            <PlanningBox points={state.tripPoints} 
                        selectedPoint={state.selectedTripPoint} 
                        onPointSelect={handleTripPointSelection} 
                        onPointDelete={handleTripPointDeletion}
                        className="row-start-3" />
            <MapBox selectedStagingPoint={state.selectedStagingPoint}
                    selectedTripPoint={state.selectedTripPoint}
                    stagingPoints={state.stagingPoints} 
                    tripPoints={state.tripPoints} 
                    onStagingPointSelection={handleStagingPointSelection}
                    onTripPointSelection={handleTripPointSelection}
                    className="row-start-4"/>
        </div>
    )
}