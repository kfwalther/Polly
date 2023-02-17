import React from "react";
import { useSortBy, useTable } from "react-table";
import { toPercent, toUSD } from "./Helpers";


export function StockTable({ data }) {

    // Define the column names and format for our Stock table.
    const columns = React.useMemo(
        () => [
            {
                Header: 'Ticker',
                accessor: 'ticker', // accessor is the "key" in the data
                sortType: 'basic',
            },
            {
                Header: 'Market Price',
                accessor: 'marketPrice',
                Cell: props => <React.Fragment>{toUSD(props.value)}</React.Fragment>,
                sortType: 'basic',
            },
            {
                Header: 'Avg Cost',
                accessor: 'unitCostBasis',
                Cell: props => <React.Fragment>{toUSD(props.value)}</React.Fragment>,
                sortType: 'basic',
            },
            {
                Header: '1D Gain %',
                accessor: 'dailyGainPercentage',
                Cell: props => <React.Fragment>{toPercent(props.value)}</React.Fragment>,
                sortType: 'basic',
            },
            {
                Header: '1D Gain',
                accessor: 'dailyGain',
                Cell: props => <React.Fragment>{toUSD(props.value)}</React.Fragment>,
                sortType: 'basic',
            },
            {
                Header: 'Unrealized Gain',
                accessor: 'unrealizedGain',
                Cell: props => <React.Fragment>{toUSD(props.value)}</React.Fragment>,
                sortType: 'basic',
            },
            {
                Header: 'Market Value',
                accessor: 'marketValue',
                Cell: props => <React.Fragment>{toUSD(props.value)}</React.Fragment>,
                sortType: 'basic',
            },
            {
                Header: 'Total Cost',
                accessor: 'totalCostBasis',
                Cell: props => <React.Fragment>{toUSD(props.value)}</React.Fragment>,
                sortType: 'basic',
            },
            {
                Header: 'Realized Gain',
                accessor: 'realizedGain',
                Cell: props => <React.Fragment>{toUSD(props.value)}</React.Fragment>,
                sortType: 'basic',
            },
            {
                Header: 'Total Gain',
                accessor: 'totalGain',
                Cell: props => <React.Fragment>{toUSD(props.value)}</React.Fragment>,
                sortType: 'basic',
            },
            {
                Header: 'Number of Shares',
                accessor: 'numShares',
                sortType: 'basic',
            },
        ], []
    );

    // Define the table, with the useTable hook.
    const {
        getTableProps, // table props from react-table
        getTableBodyProps, // table body props from react-table
        rows, // rows for the table based on the data passed
        headerGroups,
        prepareRow, // Prepare the row (this function needs to be called for each row before getting the row props)
    } = useTable({
        columns,
        data
    },
        useSortBy
    );

    // Return the react-table, with some sorting options.
    return (
        <table {...getTableProps()}>
            <thead>
                {headerGroups.map(headerGroup => (
                    <tr {...headerGroup.getHeaderGroupProps()}>
                        {headerGroup.headers.map(column => (
                            <th {...column.getHeaderProps(column.getSortByToggleProps())}>
                                {column.render("Header")}
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