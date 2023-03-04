import React from "react";
import { toPercent, toUSD, getDateFromUtcDateTime } from "./Helpers";
import { StockTable, TABLE_GREEN, TABLE_RED } from "./StockTable"

export function TransactionsTable({ txnData }) {

    // Define the column names and format for our holdings table.
    const txnCols = React.useMemo(
        () => [
            {
                Header: 'Date',
                accessor: 'dateTime',
                Cell: props => <>{getDateFromUtcDateTime(props.value)}</>,
                sortType: 'basic',
            },
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
            {
                Header: 'Price',
                accessor: 'price',
                Cell: props => <>{toUSD(props.value)}</>,
                sortType: 'basic',
            },
            {
                Header: 'Shares',
                accessor: 'shares',
                sortType: 'basic',
            },
            {
                Header: 'Value',
                accessor: 'value',
                Cell: props => <>{toUSD(props.value)}</>,
                sortType: 'basic',
            },
            {
                Header: 'Total Return %',
                accessor: 'totalReturn',
                Cell: props =>
                    <div style={{ color: props.value >= 0 ? TABLE_GREEN : TABLE_RED }} >
                        {toPercent(props.value)}
                    </div>,
                sortType: 'basic',
            },
            {
                Header: 'S&P500 Return %',
                accessor: 'sp500Return',
                Cell: props =>
                    <div style={{ color: props.value >= 0 ? TABLE_GREEN : TABLE_RED }} >
                        {toPercent(props.value)}
                    </div>,
                sortType: 'basic',
            },
            {
                Header: 'Excess Return %',
                accessor: 'excessReturn',
                Cell: props =>
                    <div style={{ color: props.value >= 0 ? TABLE_GREEN : TABLE_RED }} >
                        {toPercent(props.value)}
                    </div>,
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