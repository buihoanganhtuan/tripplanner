import { MapContainer, TileLayer, useMapEvents, Marker, Popup } from 'react-leaflet'
import { GeoPoint } from './PlanningPane'
import 'leaflet/dist/leaflet.css'
import { v4 as uuidv4 } from 'uuid';
import { icon, LeafletEventHandlerFnMap, LeafletMouseEvent } from 'leaflet'
import { Point } from './PlanningBox';
import { useEffect, useRef } from 'react';

interface MapBoxProps extends BaseComponentProps {
    selectedStagingPoint: GeoPoint | null
    selectedTripPoint: Point | null
    stagingPoints: GeoPoint[]
    tripPoints: Point[]
    onStagingPointSelection: (g: GeoPoint) => void
    onTripPointSelection: (g: Point) => void
}

export function MapBox(props: MapBoxProps) {
    let sp = props.stagingPoints.map(p => <LocationMarker key={uuidv4()} point={p} onSelection={() => props.onStagingPointSelection(p)} selected={ props.selectedStagingPoint == p } iconUrl='../../assets/black_marker.png'/>)
    let tp = props.tripPoints.map(p => <LocationMarker key={uuidv4()} point={p} onSelection={() => props.onTripPointSelection(p)} selected={ props.selectedTripPoint == p } iconUrl='../../assets/green_marker.png'/>)

    return (
        <div className={"mb-5 " + props.className}>
            <MapContainer className='h-[30rem] w-[50rem] max-h-[30rem] max-w-[50rem]' center={ { lat: 35.66669276177879, lng: 139.75811448794653 } } zoom={18} scrollWheelZoom={true}>
                <TileLayer attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors' url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png" />
                {sp}
                {tp}
            </MapContainer>
        </div>        
    )
}

interface LocationMarkerProps {
    point: GeoPoint
    selected: boolean
    onSelection: (point: GeoPoint) => void
    iconUrl: string
}

function LocationMarker(props: LocationMarkerProps) {
    let pos = { lat: props.point.lat, lng: props.point.lon }
    let addr = props.point.address
    let name = props.point.name

    const ref = useRef(null)
    // useEffect(() => {
    //     if (ref.current != null && props.selected)
    //         ref.current.openPopup()
    // })

    if (props.selected) {
        const map = useMapEvents({})
        map.flyTo(pos, map.getZoom())
    }

    let eventHandlers: LeafletEventHandlerFnMap = {
        click: (e: LeafletMouseEvent) => {
            // e.target.closePopup()
            props.onSelection(props.point)
        },
    }

    let ic = icon({
		iconUrl:       props.selected ? '../../assets/red_marker.png' : props.iconUrl,
		iconRetinaUrl: props.selected ? '../../assets/red_marker.png' : props.iconUrl,
		iconSize:    [25, 41],
		iconAnchor:  [12, 41],
		popupAnchor: [1, -34],
		tooltipAnchor: [16, -28],
		shadowSize:  [41, 41]
	})

    return (
        <Marker position={pos} eventHandlers={eventHandlers} icon={ic} ref={ref}>
            <Popup>
                {name}<br />{`${addr.prefecture}, ${addr.city}${addr.district ? `, ${addr.district}` : ''}${addr.landcode ? `, ${addr.landcode}` : ''}`}
            </Popup>
        </Marker>
    )
}