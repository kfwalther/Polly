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

    serverRequest() {
        // $.get("http://localhost:3000/securities", res => {
        //     this.setState({
        //         stockList: res
        //     });
        // });
        fetch("http://localhost:5000/securities")
            .then(response => response.json())
            .then(secs => this.setState({ stockList: secs["securities"] }));
    }

    componentDidMount() {
        this.serverRequest();
    }

    render() {
        return (
            <table>
                <thead>
                    <tr>
                        <th>1</th>
                        <th>2</th>
                        <th>3</th>
                        <th>4</th>
                        <th>5</th>
                        <th>6</th>
                        <th>7</th>
                        <th>8</th>
                    </tr>
                </thead>
                <caption>Stock List</caption>
                {this.state.stockList.map(stock => {
                    <tr key={stock.ticker}>
                        {Object.values(stock).map((val) => (
                            <td>{val}</td>
                        ))}
                    </tr>
                })}
            </table>
        )
    }
}

// <tbody>
// {this.state.stockList.map(stock => {
//     <Stock key={stock.name} stock={stock} />
// })}
// </tbody>