import React from 'react'

export default function Stock({ stock }) {
    return (
        <tr>
            <td>{stock.ticker}</td>
            <td>{stock.marketPrice}</td>
            <td>{stock.numShares}</td>
            <td>{stock.unrealizedGains}</td>
            <td>{stock.realizedGains}</td>
        </tr>
    )
}
