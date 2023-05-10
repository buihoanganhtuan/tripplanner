import { MapContainer, TileLayer, useMapEvents, Marker, Popup } from 'react-leaflet'
import { GeoPoint } from './PlanningPane'
import 'leaflet/dist/leaflet.css'
import { icon, Icon, LeafletEventHandlerFnMap, LeafletMouseEvent } from 'leaflet'

interface MapBoxProps extends BaseComponentProps {
    selectedPoint: GeoPoint | null
    stagingPoints: GeoPoint[]
    onPointSelection: (g: GeoPoint) => void
}

export function MapBox(props: MapBoxProps) {
    let sp = props.stagingPoints.map(p => <LocationMarker point={p} onSelection={() => props.onPointSelection(p)} color='blue' selected={ props.selectedPoint !== null && props.selectedPoint.id === p.id }/>)

    return (
        <div className={"mb-5 " + props.className}>
            <MapContainer className='h-[30rem] w-[50rem] max-h-[30rem] max-w-[50rem]' center={ { lat: 35.66669276177879, lng: 139.75811448794653 } } zoom={18} scrollWheelZoom={true}>
                <TileLayer attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors' url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png" />
                {sp}
            </MapContainer>
        </div>        
    )
}

interface LocationMarkerProps {
    point: GeoPoint
    onSelection: (g: GeoPoint) => void
    color: string
    selected: boolean
}

function LocationMarker(props: LocationMarkerProps) {
    let pos = { lat: props.point.lat, lng: props.point.lon }
    let addr = props.point.address
    let name = props.point.name

    if (props.selected) {
        const map = useMapEvents({})
        map.flyTo(pos, map.getZoom())
    }

    let eventHandlers: LeafletEventHandlerFnMap = {
        click: (e: LeafletMouseEvent) => {
            props.onSelection(props.point)
        }
    }

    let ic = icon({
		iconUrl:       props.selected ? '../../assets/red_marker.png' : '../../assets/black_marker.png',
		iconRetinaUrl: props.selected ? '../../assets/red_marker.png' : '../../assets/black_marker.png',
		iconSize:    [25, 41],
		iconAnchor:  [12, 41],
		popupAnchor: [1, -34],
		tooltipAnchor: [16, -28],
		shadowSize:  [41, 41]
	})

    console.log(props.point.id)
    console.log(Icon.Default.prototype.options.iconUrl)

    return (
        <Marker position={pos} eventHandlers={eventHandlers} icon={ic}>
            <Popup>
                {name}<br />{`${addr.prefecture}, ${addr.city}${addr.district ? `, ${addr.district}` : ''}${addr.landcode ? `, ${addr.landcode}` : ''}`}
            </Popup>
        </Marker>
    )
}