import { PointPlanningBox } from '../PointPlanningBox';
import { DropdownBox } from './DropdownList';
import { GeoPoint } from "./PlanningPane";
import { v4 as uuidv4 } from 'uuid';
import { Fragment } from 'react';

export interface Point extends GeoPoint {
    before: string[],
    after: string[],
    isFirst: boolean,
    isLast: boolean,
    arriveBefore: string,
    stayDuration: number
}

interface PlanningBoxProps extends BaseComponentProps {
    points: Point[]
    selectedPoint: Point | null
    onPointSelect: (point: Point) => void
    onPointDelete: (point: Point) => void
}

export function PlanningBox(props: PlanningBoxProps) {
    let compList = props.points.map(p => <Fragment key={uuidv4()}><PointPlanningBox point={p} selected={p == props.selectedPoint} onSelect={props.onPointSelect} onDelete={props.onPointDelete}/></Fragment>)

    return (
        <div className={props.className}>
            <DropdownBox name="Trip points">
                {compList}
            </DropdownBox>
        </div>
    )
}