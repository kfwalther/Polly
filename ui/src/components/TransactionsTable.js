import React from "react";
import { toPercent, toUSD } from "./Helpers";
import { StockTable, TABLE_GREEN, TABLE_RED } from "./StockTable"

export function TransactionsTable({ txnData }) {

    // Define the column names and format for our holdings table.
    const txnCols = React.useMemo(
        () => [
            {
                Header: 'Ticker',
                accessor: 'ticker', // accessor is the "key" in the data
                Cell: props =>
                    <div style={{ fontWeight: 'bold' }}>
                        {props.value}
                    </div>,
                sortType: 'basic',
            },
            {
                Header: 'Action',
                accessor: 'action',
                sortType: 'basic',
            },
        ], []
    );

    // Return the react-table, with some sorting options.
    return (
        <StockTable
            data={txnData}
            columns={txnCols}
        />
    );
}