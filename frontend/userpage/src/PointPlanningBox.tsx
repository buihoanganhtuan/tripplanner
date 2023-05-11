import React, { useState } from "react"
import { Point } from "./components/PlanningBox"
import { v4 as uuidv4 } from 'uuid';

interface PointPlanningBoxProps extends BaseComponentProps {
    point: Point
    selected: boolean
    onSelect: (point: Point) => void
    onDelete: (point: Point) => void
}

interface PointPlanningBoxLocalState {
    collapsed: boolean
}

export function PointPlanningBox(props: PointPlanningBoxProps) {
    const [state, setState] = useState<PointPlanningBoxLocalState>({
        collapsed: false
    })

    const handleExpand = (e: React.MouseEvent<HTMLButtonElement>) => {
        e.stopPropagation()
        setState(prev => { return {...prev, collapsed: !prev.collapsed} })
    }

    const reportDelete = (e: React.MouseEvent<HTMLButtonElement>) => {
        e.stopPropagation()
        props.onDelete(props.point)
    }

    const reportSelect = () => {
        props.onSelect(props.point)
    }

    let p = props.point
    return (
    <div onClick={reportSelect} className={"grid grid-cols-[1fr_0.3fr_0.1fr] auto-rows-min border-2 rounded-md " + (props.selected ? "bg-amber-200 " : " ") + props.className}>
        <div className="col-start-1 row-start-1 pl-[6px] pb-[8px] pt-[8px]">
            {p.name}, Address: {`${p.address.prefecture}, ${p.address.city} city${p.address.district == null ? '' : ', ' + p.address.district}${p.address.landcode == null ? '' : ', ' + p.address.landcode}`}
        </div>
        <button onClick={handleExpand} className="col-start-2 row-start-1">{state.collapsed ? "Hide constraints" : "Show constraints"}</button>
        <button onClick={reportDelete} className="col-start-3 row-start-1">X</button>
        <div className={"row-start-2 col-start-1 col-end-3 transition-[max-height] duration-200 overflow-hidden " + (state.collapsed ? "max-h-[1080px]" : "max-h-0")}>
            <div>{"Before: " + props.point.before}</div>
            <div>{"After: " + props.point.after}</div>
            <div>{"Is first point: " + props.point.isFirst}</div>
            <div>{"Is last point: " + props.point.isLast}</div>
            <div>{"Arrive by: " + props.point.arriveBefore}</div>
            <div>{"Time spend: " + props.point.stayDuration}</div>
        </div>
    </div>
    )
}