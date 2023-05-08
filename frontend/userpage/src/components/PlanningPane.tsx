import React, { useState } from "react"
import { AutocompleteBox } from "./AutocompleteBox"
import { SearchBox } from "./SearchBox"
import { StagingBox } from "./StagingBox"

interface PlanningPaneState {
    stagingBoxPoints: GeoPoint[]
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
}

export function PlanningPane() {
    const [state, setState] = useState<PlanningPaneState>({
        stagingBoxPoints: [],
        autocompletedPoint: null
    })

    const handleAutocompletion = (g: GeoPoint) => {
        setState(prev => {
            prev.stagingBoxPoints.push(g)
            return prev
        })
    }

    const handlePointDeletion = (list: GeoPoint[]) => {
        setState(prev => {
            return { ...prev, stagingBoxPoints: list }
        })
    }

    return (
        <div>
            <SearchBox input="" selectedEntry={state.autocompletedPoint} onEntrySelection={handleAutocompletion}/>
            <StagingBox points={state.stagingBoxPoints} onDelete={handlePointDeletion}/>
        </div>
    )
}