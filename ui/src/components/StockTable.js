import React from "react";
import { useFilters, useSortBy, useTable } from "react-table";
import { toUSD } from "./Helpers";

// Much of the code in this file was taken from the react-table examples:
// https://github.com/TanStack/table/blob/v7/examples/filtering/src/App.js

// This is a custom filter UI for selecting
// a unique option from a list
function SelectColumnFilter({
    column: { filterValue, setFilter, preFilteredRows, id },
}) {
    // Calculate the options for filtering
    // using the preFilteredRows
    const options = React.useMemo(() => {
        const options = new Set()
        preFilteredRows.forEach(row => {
            options.add(parseFloat(row.values[id]).toFixed(2))
        })
        return [...options.values()]
    }, [id, preFilteredRows])

    // Render a multi-select box
    return (
        <select
            value={filterValue}
            onChange={e => {
                setFilter(e.target.value || undefined)
            }}
        >
            <option value="">All</option>
            {options.map((option, i) => (
                <option key={i} value={option}>
                    {option}
                </option>
            ))}
        </select>
    )
}

// Define a custom filter filter function, to filter out selected values.
function filterNotEqualTo(rows, id, filterValue) {
    return rows.filter(row => {
        const rowValue = row.values[id]
        return rowValue !== filterValue.toString()
    })
}

export function StockTable({ data }) {

    // Define the column names and format for our Stock table.
    const columns = React.useMemo(
        () => [
            {
                Header: 'Ticker',
                accessor: 'ticker', // accessor is the "key" in the data
                disableFilters: true,
                sortType: 'basic',
            },
            {
                Header: 'Market Price',
                accessor: 'marketPrice',
                Cell: props => <React.Fragment>{toUSD(props.value)}</React.Fragment>,
                disableFilters: true,
                sortType: 'basic',
            },
            {
                Header: 'Market Value',
                accessor: 'marketValue',
                Filter: SelectColumnFilter,
                filter: filterNotEqualTo,
                filterValue: 0.0,
                sortType: 'basic',
                Cell: props => <React.Fragment>{toUSD(props.value)}</React.Fragment>,
            },
            {
                Header: 'Unit Cost Basis',
                accessor: 'unitCostBasis',
                Cell: props => <React.Fragment>{toUSD(props.value)}</React.Fragment>,
                disableFilters: true,
                sortType: 'basic',
            },
            {
                Header: 'Total Cost Basis',
                accessor: 'totalCostBasis',
                Cell: props => <React.Fragment>{toUSD(props.value)}</React.Fragment>,
                disableFilters: true,
                sortType: 'basic',
            },
            {
                Header: 'Number of Shares',
                accessor: 'numShares',
                disableFilters: true,
                sortType: 'basic',
            },
            {
                Header: 'Realized Gains',
                accessor: 'realizedGains',
                Cell: props => <React.Fragment>{toUSD(props.value)}</React.Fragment>,
                disableFilters: true,
                sortType: 'basic',
            },
            {
                Header: 'Unrealized Gains',
                accessor: 'unrealizedGains',
                Cell: props => <React.Fragment>{toUSD(props.value)}</React.Fragment>,
                disableFilters: true,
                sortType: 'basic',
            },
        ], []
    );

    // Let's set up our default Filter UI.
    const defaultColumn = React.useMemo(
        () => ({
            Filter: SelectColumnFilter,
        }),
        []
    )

    // Define the table, with the useTable hook.
    const {
        getTableProps, // table props from react-table
        getTableBodyProps, // table body props from react-table
        rows, // rows for the table based on the data passed
        headerGroups,
        prepareRow, // Prepare the row (this function needs to be called for each row before getting the row props)
    } = useTable({
        columns,
        data,
        defaultColumn
    },
        useFilters,
        useSortBy
    );

    // Return the react-table, with some filtering options.
    return (
        <table {...getTableProps()}>
            <thead>
                {headerGroups.map(headerGroup => (
                    <tr {...headerGroup.getHeaderGroupProps()}>
                        {headerGroup.headers.map(column => (
                            <th {...column.getHeaderProps(column.getSortByToggleProps())}>
                                {column.render("Header")}
                                <div>
                                    {column.canFilter ? column.render("Filter") : null}
                                </div>
                                <span>
                                    {column.isSorted ? (column.isSortedDesc ? ' 🔽' : ' 🔼') : ''}
                                </span>
                            </th>
                        ))}
                    </tr>
                ))}
            </thead>
            <tbody {...getTableBodyProps()}>
                {rows.map((row, i) => {
                    prepareRow(row);
                    return (
                        <tr {...row.getRowProps()}>
                            {row.cells.map(cell => {
                                return <td {...cell.getCellProps()}>{cell.render("Cell")}</td>;
                            })}
                        </tr>
                    );
                })}
            </tbody>
        </table>
    );
}