import React from 'react'
import Stock from './Stock'

export default function StockList({ stockList }) {
    return (
        stockList.map(stock => {
            return <Stock key={stock.name} stock={stock} />
        })
    )
}
