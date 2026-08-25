'''
Python script that uses yfinance to extract the encrypted data stores from the Yahoo finance web page. It does this using 
the BeautifulSoup web scraper.

Tested with Python:
3.11.3
pip install yfinance --upgrade --no-cache-dir

If any issues, which are frequent, check here:
https://github.com/ranaroussi/yfinance/issues

Sometimes, you just need the latest version of the module:
pip install --upgrade yfinance

If errors running this script, it's likely that an old ticker got delisted from exchange and no longer queryable.
'''

import json
import sys
import pandas as pd
import yfinance as yf

def get_extended_info(symbols):
    """Get extended information for multiple stock symbols"""
    if not symbols or symbols.isspace():
        return None
        
    # Put spaces b/w symbols instead of commas
    symbols_spaced = ' '.join(symbols.split(','))
    # Get the info for all symbols
    stock_info_dict = yf.Tickers(symbols_spaced)
    # Use dict-comprehension to pull out the 'info' portion
    reduced_stock_info_dict = {ticker: stock_info.info for ticker, stock_info in stock_info_dict.tickers.items()}
    return reduced_stock_info_dict

def get_historical_data(ticker, start_date, end_date):
    """Get historical price data for a single ticker"""
    if not ticker or ticker.isspace():
        return None
        
    # Get historical data
    stock = yf.Ticker(ticker)
    history = stock.history(start=start_date, end=end_date, interval="1d")
    
    # Convert to dict format matching Quote structure
    history_dict = {
        "symbol": ticker,
        "date": history.index.strftime('%Y-%m-%dT%H:%M:%SZ').tolist(),
        "open": history['Open'].tolist(),
        "high": history['High'].tolist(),
        "low": history['Low'].tolist(),
        "close": history['Close'].tolist(),
        "volume": history['Volume'].tolist()
    }
    return history_dict


def _symbol_frame(history, symbol):
    """Extract one symbol's columns from yfinance's grouped DataFrame."""
    if isinstance(history.columns, pd.MultiIndex):
        for level in range(history.columns.nlevels):
            if symbol in history.columns.get_level_values(level):
                return history.xs(symbol, axis=1, level=level, drop_level=True)
        return None
    return history


def _finite_values(frame, column):
    if column not in frame:
        return [0.0] * len(frame)
    return [0.0 if pd.isna(value) else float(value) for value in frame[column].tolist()]


def get_historical_data_batch(symbols, start_date, end_date):
    """Download daily historical data for several symbols in one request."""
    symbols = [symbol.strip() for symbol in symbols if symbol and symbol.strip()]
    if not symbols:
        return {}

    history = yf.download(
        tickers=symbols,
        start=start_date,
        end=end_date,
        interval="1d",
        group_by="ticker",
        threads=False,
        progress=False,
        auto_adjust=True,
    )
    result = {}
    for symbol in symbols:
        frame = _symbol_frame(history, symbol)
        if frame is None or "Close" not in frame:
            continue
        frame = frame.dropna(subset=["Close"])
        result[symbol] = {
            "symbol": symbol,
            "date": frame.index.strftime('%Y-%m-%dT%H:%M:%SZ').tolist(),
            "open": _finite_values(frame, "Open"),
            "high": _finite_values(frame, "High"),
            "low": _finite_values(frame, "Low"),
            "close": _finite_values(frame, "Close"),
            "volume": _finite_values(frame, "Volume"),
        }
    return result


# Main entry point to the script
if __name__ == "__main__":
    # Parse command line args
    args = sys.argv[1:]

    if len(args) == 1:
        # Extended info mode
        result = get_extended_info(args[0])
    elif len(args) == 3:
        # Historical data mode
        result = get_historical_data(args[0], args[1], args[2])
    elif len(args) == 4 and args[0] == "--historical-batch":
        result = get_historical_data_batch(args[1].split(','), args[2], args[3])
    else:
        print("Invalid arguments")
        sys.exit(1)

    # Output results as JSON
    print(json.dumps(result, indent=4, allow_nan=False))
