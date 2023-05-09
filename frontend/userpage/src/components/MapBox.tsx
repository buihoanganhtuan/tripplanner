import { MapContainer, TileLayer, useMapEvents, Marker, Popup } from 'react-leaflet'
import { GeoPoint } from './PlanningPane'
import 'leaflet/dist/leaflet.css'

interface MapBoxProps {
    selectedPoint: GeoPoint | null
}

export function MapBox(props: MapBoxProps) {
    return (
        <div>
            <MapContainer className='h-[30rem] w-[50rem]' center={ { lat: 35.66669276177879, lng: 139.75811448794653 } } zoom={18} scrollWheelZoom={true}>
                <TileLayer attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors' url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png" />
                <LocationMarker point={props.selectedPoint} />
            </MapContainer>
        </div>
    )
}

interface LocationMarkerProps {
    point: GeoPoint | null
}

function LocationMarker(props: LocationMarkerProps) {
    if (props.point == null)
        return null

    let pos = { lat: props.point.lat, lng: props.point.lon }
    let addr = props.point.address
    let name = props.point.name

    const map = useMapEvents({})
    map.flyTo(pos, map.getZoom())

    return (
        <Marker position={pos}>
            <Popup>
                {name}<br />{`${addr.prefecture}, ${addr.city}${addr.district ? `, ${addr.district}` : ''}${addr.landcode ? `, ${addr.landcode}` : ''}`}
            </Popup>
        </Marker>
    )
}