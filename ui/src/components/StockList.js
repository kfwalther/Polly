import React from 'react'
import Stock from './Stock'

export default class StockList extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            stockList: []
        };
        this.serverRequest = this.serverRequest.bind(this);
    }

    // Fetch the stock list from the server.
    serverRequest() {
        fetch("http://localhost:5000/securities")
            .then(response => response.json())
            .then(secs => this.setState({ stockList: secs["securities"] }));
    }

    // Runs on component mount, to grab data from the server.
    componentDidMount() {
        this.serverRequest();
    }

    // Returns the HTML to display the stock table.
    renderStockTable() {
        return (
            <table>
                <caption>Stock List</caption>
                <thead>
                    <tr>
                        {Object.keys(this.state.stockList[0]).map((header) => (
                            <th>{header}</th>
                        ))}
                    </tr>
                </thead>
                <tbody>
                    {this.state.stockList.map(stock => {
                        return <tr key={stock.ticker}>
                            {Object.values(stock).map((val) => (
                                <td>{val}</td>
                            ))}
                        </tr>
                    })}
                </tbody>
            </table>
        )
    }

    // Render the stock table, or a loader screen until data is retrieved from server.
    render() {
        const curState = this.state
        return curState.stockList.length ? this.renderStockTable() : (
            <span>LOADING STOCKS...</span>
        )
    }
}

// <tbody>
// {this.state.stockList.map(stock => {
//     <Stock key={stock.name} stock={stock} />
// })}
// </tbody>

