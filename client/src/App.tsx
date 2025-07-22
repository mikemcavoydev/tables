import { useEffect, useState, type FormEventHandler } from "react";
import "./App.css";
import { Env } from "./Env";

type Sorting = { column: string; order: string };

type Table = {
    id: number;
    title: string;
    items: Item[];
};

type Item = {
    id: number;
    title: string;
    tags?: Tag[];
};

type Tag = {
    id: number;
    title: string;
    description: string;
};

type SortFn = (props: { column: string; order: string }) => void;

interface HeaderCellProps {
    column: string;
    sorting: Sorting;
    sortFn: SortFn;
}

function HeaderCell({ column, sorting, sortFn }: HeaderCellProps) {
    const isDescSorting = sorting.column === column && sorting.order === "desc";
    const isAscSorting = sorting.column === column && sorting.order === "asc";
    const nextSortingOrder = isAscSorting ? "desc" : "asc";

    return (
        <th
            className="table-cell"
            onClick={() => sortFn({ column, order: nextSortingOrder })}
        >
            {column}
            {isDescSorting && <span>▾</span>}
            {isAscSorting && <span>▴</span>}
        </th>
    );
}

interface HeaderProps {
    columns: string[];
    sorting: Sorting;
    sortFn: SortFn;
}

function Header({ columns, sorting, sortFn }: HeaderProps) {
    return (
        <thead>
            <tr>
                {columns.map((col) => (
                    <HeaderCell
                        key={col}
                        column={col}
                        sorting={sorting}
                        sortFn={sortFn}
                    />
                ))}
            </tr>
        </thead>
    );
}

interface BodyProps {
    entries: Item[];
    columns: string[];
}

function Body({ entries, columns }: BodyProps) {
    return (
        <tbody>
            {entries.map((entry) => (
                <tr key={entry.id}>
                    {columns.map((col) => (
                        <td key={col} className="table-cell">
                            {entry[col]}
                        </td>
                    ))}
                </tr>
            ))}
        </tbody>
    );
}

type SearchFn = (value: string) => void;

interface SearchBarProps {
    searchFn: SearchFn;
}

function SearchBar({ searchFn }: SearchBarProps) {
    const [searchValue, setSearchValue] = useState("");
    const submit: FormEventHandler<HTMLFormElement> = (e) => {
        e.preventDefault();
        searchFn(searchValue);
    };

    return (
        <div className="search-bar">
            <form onSubmit={submit}>
                <input
                    type="text"
                    placeholder="Search..."
                    value={searchValue}
                    onChange={(e) => setSearchValue(e.target.value)}
                />
            </form>
        </div>
    );
}

interface TableProps {
    data: Table;
}

function Table({ data }: TableProps) {
    const [sorting, setSorting] = useState<Sorting>({
        column: "id",
        order: "asc",
    });

    const [searchValue, setSearchValue] = useState("");

    const sortFn: SortFn = (newSort) => {
        setSorting(newSort);
    };

    const searchFn = (value: string) => {
        setSearchValue(value);
    };

    const columns = Object.keys(data.items[0]).filter((key) => key !== "tags");

    return (
        <div>
            <SearchBar searchFn={searchFn} />
            <table className="table">
                <Header columns={columns} sorting={sorting} sortFn={sortFn} />
                <Body columns={columns} entries={data.items} />
            </table>
        </div>
    );
}

function App() {
    const [data, setData] = useState<Table[]>([]);

    useEffect(() => {
        fetch(`${Env.API_BASE_URL}/tables`)
            .then((res) => res.json())
            .then((data) => setData(data.data));
    }, []);

    return <div>{data.length > 0 ? <Table data={data[0]} /> : null}</div>;
}

export default App;
