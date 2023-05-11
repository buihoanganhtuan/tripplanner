import { useLayoutEffect, useRef, useState } from "react"

interface DropdownBoxProps<T extends JSX.Element> extends BaseComponentProps {
    name: string
    children: T[]
}

interface DropdownBoxLocalState {
    collapsed: boolean
}

export function DropdownBox<T extends JSX.Element>(props: DropdownBoxProps<T>) {
    const [state, setState] = useState<DropdownBoxLocalState>({
        collapsed: false,
    })

    const handleBarClick = () => {
        setState(prev => { 
            return { ...prev, collapsed: !prev.collapsed } 
        })
    }

    // https://stackoverflow.com/questions/47977751/css-set-parents-height-to-0-but-child-div-still-show
    return (
        <div className="grid grid-rows-drop-down justify-items-center items-center gap-y-2">
            <div onMouseDown={handleBarClick} className="row-start-1 border-2 rounded-md w-[20rem] h-[30px] bg-green-200 drop-shadow-lg grid justify-items-center ">{`${props.name} (${props.children.length})`}</div>
            <div className={"grid row-start-2 bg-white rounded-lg w-[50rem] transition-[max-height] duration-400ms overflow-hidden " + (state.collapsed && props.children.length > 0 ? "max-h-[1080px] border-1" : "max-h-0 border-0")}>
                {props.children}
            </div>
        </div>
    )
}