import { v4 as uuidv4 } from 'uuid';
import { GeoPoint } from './PlanningPane';

interface StagingBoxProps {
    points: GeoPoint[]
    onDelete: (points: GeoPoint[]) => void
}

export function StagingBox(props: StagingBoxProps) {
    const handleDeletion = (id: string) => {
        props.onDelete(props.points.filter(p => p.id !== id))
    }

    let list = props.points.map(p =>
        <div key={uuidv4()}>
            Point name: {p.name}, address: {`${p.address.prefecture}, ${p.address.city} city, ${p.address.district}${p.address.landcode == null ? '' : ' , ' + p.address.landcode}`}
            <button onClick={() => handleDeletion(p.id)}>X</button>
        </div>)

    return (
        <div>{list}</div>
    )
}