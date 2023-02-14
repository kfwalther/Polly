import React from 'react'
import StockTable from './StockTable'

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

    // Returns the HTML to display the stock table.
    renderStockTable() {
        return (
            <StockTable data={this.state.stockList} />
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
