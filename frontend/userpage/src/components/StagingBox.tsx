import React from 'react';
import { v4 as uuidv4 } from 'uuid';
import { DropdownBox } from './DropdownList';
import { GeoPoint } from './PlanningPane';

interface StagingBoxProps extends BaseComponentProps {
    points: GeoPoint[]
    selectedPoint: GeoPoint | null
    onPointDelete: (point: GeoPoint) => void
    onPointSelect: (point: GeoPoint) => void
    onPointTransfer: (point: GeoPoint) => void
}

export function StagingBox(props: StagingBoxProps) {
    const handleDeletion = (e: React.MouseEvent<HTMLButtonElement>, gp: GeoPoint) => {
        e.stopPropagation()
        props.onPointDelete(gp)
    }

    const handleTransfer = (e: React.MouseEvent<HTMLButtonElement>, gp: GeoPoint) => {
        e.stopPropagation()
        props.onPointTransfer(gp)
    }

    let list = props.points.map(p => {
        let style = props.selectedPoint !== null && props.selectedPoint.id === p.id ? " bg-amber-200 " : ""
        return (
            <div key={uuidv4()} onClick={() => props.onPointSelect(p)} className={style + " grid grid-cols-[1fr_0.3fr_0.1fr] border-2 rounded-md"}>
                <div className="col-start-1 pl-[6px] pb-[8px] pt-[8px]">
                    {p.name}, Address: {`${p.address.prefecture}, ${p.address.city} city${p.address.district == null ? '' : ', ' + p.address.district}${p.address.landcode == null ? '' : ', ' + p.address.landcode}`}
                </div>
                <button onClick={e => handleTransfer(e, p)} className="col-start-2">Add to trip</button>
                <button onClick={e => handleDeletion(e, p)} className="col-start-3">X</button>
            </div>
        )
    })

    return (
        <div className={props.className}>
            <DropdownBox name="Selected points" children={list} />
        </div>
    )
}