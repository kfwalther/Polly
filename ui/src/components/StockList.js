import React from 'react'

export default class StockList extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            stockList: []
        };
        this.serverRequest = this.serverRequest.bind(this);
        // Assign this instance to a global variable.
        window.stockList = this;
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

    applySorting(sortType) {
        console.log("sorting by " + sortType.col);
        // Apply sort setting to this local copy of stocks.
        const sortedList = [...this.state.stockList].sort((a, b) => {
            // Check if sorting alphabetically or numerically.
            if (sortType.col === "ticker") {
                return (sortType.ascending ? 1 : -1) * a[sortType.col].localeCompare(b[sortType.col]);
            } else {
                return (sortType.ascending ? 1 : -1) * a[sortType.col] - b[sortType.col];
            }
        });
        // Save the sorted list of stocks.
        this.setState({ stockList: sortedList });
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
                            {Object.entries(stock).map(([k, val]) => (
                                (k == "ticker" || k == "numShares") ?
                                    <td>{val}</td> :
                                    <td><span class="dollars">{val}</span></td>
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

