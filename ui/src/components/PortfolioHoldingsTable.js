import React from "react";
import { toPercent, toUSD } from "./Helpers";
import { StockTable, TABLE_GREEN, TABLE_RED } from "./StockTable"

export function PortfolioHoldingsTable({ holdingsData, totalPortfolioValue }) {

    // Define the column names and format for our holdings table.
    const holdingsCols = React.useMemo(
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
                Header: 'Market Price',
                accessor: 'marketPrice',
                Cell: props => <>{toUSD(props.value)}</>,
                sortType: 'basic',
            },
            {
                Header: 'P/S (TTM)',
                accessor: 'priceToSalesTtm',
                Cell: props => <>{props.value !== 0.0 ? parseFloat(props.value.toFixed(3)) : '----'}</>,
                sortType: 'basic',
            },
            {
                Header: 'Rev Growth % (YoY)',
                accessor: 'revenueGrowthPercentageYoy',
                Cell: props =>
                    <div style={{ color: (props.value > 0.3) ? TABLE_GREEN : ((props.value < 0.0) ? TABLE_RED : 'white') }} >
                        {props.value !== 0.0 ? toPercent(props.value * 100) : '----'}
                    </div>,
                sortType: 'basic',
            },
            {
                Header: 'Gross Margin %',
                accessor: 'grossMargin',
                Cell: props =>
                    <div style={{ color: (props.value > 0.7) ? TABLE_GREEN : ((props.value < .4) ? TABLE_RED : 'white') }} >
                        {props.value !== 0.0 ? toPercent(props.value * 100) : '----'}
                    </div>,
                sortType: 'basic',
            },
            {
                Header: 'Avg Cost',
                accessor: 'unitCostBasis',
                Cell: props => <>{toUSD(props.value)}</>,
                sortType: 'basic',
            },
            {
                id: 'allocation',
                Header: 'Allocation %',
                accessor: 'marketValue',
                Cell: props => <>{toPercent(props.value / totalPortfolioValue * 100)}</>,
                sortType: 'basic',
            },
            {
                Header: '1D Gain',
                accessor: 'dailyGain',
                Cell: props =>
                    <div style={{ color: props.value >= 0 ? TABLE_GREEN : TABLE_RED }} >
                        {toUSD(props.value)}
                    </div>,
                sortType: 'basic',
            },
            {
                Header: '1D Gain %',
                accessor: 'dailyGainPercentage',
                Cell: props =>
                    <div style={{ color: props.value >= 0 ? TABLE_GREEN : TABLE_RED }} >
                        {toPercent(props.value)}
                    </div>,
                sortType: 'basic',
            },
            {
                Header: 'Unrealized Gain',
                accessor: 'unrealizedGain',
                Cell: props =>
                    <div style={{ color: props.value >= 0 ? TABLE_GREEN : TABLE_RED }} >
                        {toUSD(props.value)}
                    </div>,
                sortType: 'basic',
            },
            {
                Header: 'Unrealized Gain %',
                accessor: 'unrealizedGainPercentage',
                Cell: props =>
                    <div style={{ color: props.value >= 0 ? TABLE_GREEN : TABLE_RED }} >
                        {toPercent(props.value)}
                    </div>,
                sortType: 'basic',
            },
            {
                id: 'marketValue',
                Header: 'Market Value',
                accessor: 'marketValue',
                Cell: props => <>{toUSD(props.value)}</>,
                sortType: 'basic',
            },
            {
                Header: 'Total Cost',
                accessor: 'totalCostBasis',
                Cell: props => <>{toUSD(props.value)}</>,
                sortType: 'basic',
            },
            {
                Header: 'Realized Gain',
                accessor: 'realizedGain',
                Cell: props =>
                    <div style={{ color: (props.value > 0) ? TABLE_GREEN : ((props.value < 0) ? TABLE_RED : 'white') }} >
                        {props.value !== 0.0 ? toUSD(props.value) : '----'}
                    </div>,
                sortType: 'basic',
            },
            {
                Header: 'Total Gain',
                accessor: 'totalGain',
                Cell: props =>
                    <div style={{ color: props.value >= 0 ? TABLE_GREEN : TABLE_RED }} >
                        {toUSD(props.value)}
                    </div>,
                sortType: 'basic',
            },
            {
                Header: 'Number of Shares',
                accessor: 'numShares',
                Cell: props => <>{parseFloat(props.value.toFixed(3))}</>,
                sortType: 'basic',
            },
        ], []
    );

    // Return the react-table, with some sorting options.
    return (
        <StockTable
            data={holdingsData}
            columns={holdingsCols}
            initialSortCol={'ticker'}
        />
    );
}