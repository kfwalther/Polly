import React from 'react'

export default function Stock({ stock }) {
    return (
        <div>
            <label>
                {stock.name}
                {stock.price}
            </label>
        </div>
    )
}
