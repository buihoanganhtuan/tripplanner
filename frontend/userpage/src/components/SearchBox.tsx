import React, { useState } from "react"
import { v4 as uuidv4 } from 'uuid';
import { GeoPoint } from "./PlanningPane";

interface SearchBoxLocalState {
    inputText: string
    focused: boolean
}

interface SearchBoxProps extends AutocompleteBoxProps {
}

export function SearchBox(props: SearchBoxProps) {
    const [state, setState] = useState<SearchBoxLocalState>({
        inputText: '',
        focused: false
    })

    const handleInput = (e: React.ChangeEvent<HTMLInputElement>) => {
        setState(prev => { return { ...prev, inputText: e.target.value} })
        // Consult autocomplete endpoint
    }
    const handleFocus = () => setState(prev => { return { ...prev, focused: true } })
    const handleUnfocus = () => setState(prev => { return { ...prev, focused: false } })
    
    // https://stackoverflow.com/questions/10487292/position-absolute-but-relative-to-parent
    return (
        <div className={"grid grid-rows-[20px_30px_1fr] gap-y-1 relative"}>
            <label htmlFor="searchInput" className="row-start-1">Search for points</label>
            <input id="searchInput" value={state.inputText} onChange={handleInput} onFocus={handleFocus} onBlur={handleUnfocus} className={"w-[32rem] row-start-2 border-2 rounded-md"}/>
            { state.focused ?<AutocompleteBox input={state.inputText} selectedEntry={props.selectedEntry} onEntrySelection={props.onEntrySelection} className="row-start-3 absolute z-10 top-0 bg-white w-full border-2 rounded-md"/> : null}
        </div>
    )
}


interface AutocompleteBoxProps extends BaseComponentProps {
    input: string
    selectedEntry: GeoPoint | null,
    onEntrySelection: (p: GeoPoint) => void
}

const repo = ['a', 'ab', 'abc']

function AutocompleteBox(props: AutocompleteBoxProps) {
    const handleSelection = (id: string) => {
        console.log("select " + id)
        props.onEntrySelection({
            id: id,
            name: "Example id: " + id,
            address: {
                prefecture: "Tokyo",
                city: "Tokyo"
            },
            lat: 35.587218220913435 + Math.random()*0.01,
            lon: 139.72406056317507 + Math.random()*0.01,
        })
    }

    if (props.input.length == 0)
        return null
    
    // https://stackoverflow.com/questions/17769005/onclick-and-onblur-ordering-issue
    let list = repo.filter(s => s.startsWith(props.input)).map(s => {
        let k = uuidv4()
        return <tr key={k} onMouseDown={() => handleSelection(k)}><td>{s}</td></tr>
    })

    return (<table className={props.className}><tbody>{list}</tbody></table>)
}