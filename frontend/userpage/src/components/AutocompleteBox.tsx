import { v4 as uuidv4 } from 'uuid';
import { GeoPoint } from './PlanningPane';

export interface AutocompleteBoxProps {
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
            }
        })
    }

    if (props.input.length == 0)
        return null
    let list = repo.filter(s => s.startsWith(props.input)).map(s => {
        let k = uuidv4()
        return <tr key={k} onClick={() => handleSelection(k)}><td>{s}</td></tr>
    })

    return (<table><tbody>{list}</tbody></table>)
}