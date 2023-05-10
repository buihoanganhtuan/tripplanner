import { v4 as uuidv4 } from 'uuid';
import { DropdownBox } from './DropdownList';
import { GeoPoint } from './PlanningPane';

interface StagingBoxProps extends BaseComponentProps {
    points: GeoPoint[]
    selectedPoint: GeoPoint | null
    onPointDelete: (points: GeoPoint[]) => void
    onPointSelect: (point: GeoPoint) => void
}

export function StagingBox(props: StagingBoxProps) {
    const handleDeletion = (e: React.MouseEvent<HTMLButtonElement>, gp: GeoPoint) => {
        e.stopPropagation()
        props.onPointDelete(props.points.filter(p => p.id !== gp.id))
    }

    let list = props.points.map(p => {
        let style = props.selectedPoint !== null && props.selectedPoint.id === p.id ? "bg-amber-200" : ""
        return (
            <div key={uuidv4()} onClick={() => props.onPointSelect(p)} className={style}>
                Point name: {p.name}, address: {`${p.address.prefecture}, ${p.address.city} city${p.address.district == null ? '' : ', ' + p.address.district}${p.address.landcode == null ? '' : ', ' + p.address.landcode}`}
                <button onClick={e => handleDeletion(e, p)}>X</button>
            </div>
        )
    })

    return (
        <DropdownBox name="Selected points" children={list} />
    )
}