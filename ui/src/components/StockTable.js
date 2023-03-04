import React from "react";
import { useSortBy, useTable } from "react-table";

// Define the red/green colors to use in our tables.
export const TABLE_RED = '#ff2e1f'
export const TABLE_GREEN = '#56DC28'


export function StockTable({ data, columns }) {

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