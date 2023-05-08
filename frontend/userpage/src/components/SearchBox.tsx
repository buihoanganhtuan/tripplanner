import React, { useState } from "react"
import { AutocompleteBox, AutocompleteBoxProps } from "./AutocompleteBox"

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
    const handleUnfocus = () => setState(prev => { 
        console.log("Unfocused")
        return { ...prev, focused: false } 
    })
    // const handleUnfocus = () => {
    //     console.log("Unfocused")
    //     setState({ ...state, focused: false })
    // }

    let inputComp = <input value={state.inputText} onChange={handleInput} onFocus={handleFocus} onBlur={handleUnfocus}/>
    

    if (state.focused)
        return (
            <div>
                {inputComp}
                <AutocompleteBox input={state.inputText} selectedEntry={props.selectedEntry} onEntrySelection={props.onEntrySelection}/>
            </div>
        )
    
    return (<div>{inputComp}</div>)
}