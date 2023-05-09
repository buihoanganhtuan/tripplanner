import React, { useState } from "react"
import { MapBox } from "./MapBox"
import { SearchBox } from "./SearchBox"
import { StagingBox } from "./StagingBox"

interface PlanningPaneState {
    stagingPoints: GeoPoint[]
    selectedStagingPoint: GeoPoint | null
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
        selectedStagingPoint: null,
        autocompletedPoint: null
    })

    const handleStagingPointSelection = (g: GeoPoint) => {
        setState(prev => {
            return { ...prev, selectedStagingPoint: g }
        })
    }

    const handleAutocompletion = (g: GeoPoint) => {
        for (let p of state.stagingPoints) {
            if (p.id == g.id)
                return
        }
        setState(prev => {
            prev.stagingPoints.push(g)
            prev.stagingPoints.sort((p1, p2) => p1.name > p2.name ? 1 : -1)
            return { ...prev }
        })
    }

    const handlePointDeletion = (list: GeoPoint[]) => {
        for (let p of list) {
            if (p.id != state.selectedStagingPoint?.id)
                continue
            setState(prev => {
                return { ...prev, stagingPoints: list }
            })
            return
        }
        setState(prev => {
            return { ...prev, stagingPoints: list, selectedStagingPoint: null }
        })
    }

    return (
        <div className="grid grid-row-4 justify-items-center gap-y-4">
            <SearchBox input="" selectedEntry={state.autocompletedPoint} onEntrySelection={handleAutocompletion} className="min-w-[50%] border-2 rounded-md"/>
            <StagingBox points={state.stagingPoints} onPointDelete={handlePointDeletion} selectedPoint={state.selectedStagingPoint} onPointSelect={handleStagingPointSelection}/>
            <MapBox selectedPoint={state.selectedStagingPoint}/>
        </div>
    )
}