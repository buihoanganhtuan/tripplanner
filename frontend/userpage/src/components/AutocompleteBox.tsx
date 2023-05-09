import { v4 as uuidv4 } from 'uuid';
import { GeoPoint } from './PlanningPane';

export interface AutocompleteBoxProps extends BaseComponentProps {
    input: string
    selectedEntry: GeoPoint | null,
    onEntrySelection: (p: GeoPoint) => void
}

const repo = ['a', 'ab', 'abc']

export function AutocompleteBox(props: AutocompleteBoxProps) {
    const handleSelection = (id: string) => {
        console.log("select " + id)
        props.onEntrySelection({
            id: id,
            name: "Example id: " + id,
            address: {
                prefecture: "Tokyo",
                city: "Tokyo"
            },
            lat: 35.587218220913435,
            lon: 139.72406056317507,
        })
    }

    if (props.input.length == 0)
        return null
    
    // https://stackoverflow.com/questions/17769005/onclick-and-onblur-ordering-issue
    let list = repo.filter(s => s.startsWith(props.input)).map(s => {
        let k = uuidv4()
        return <tr key={k} onMouseDown={() => handleSelection(k)}><td>{s}</td></tr>
    })

    return (<table><tbody>{list}</tbody></table>)
}