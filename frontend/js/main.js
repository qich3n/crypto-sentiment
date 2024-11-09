//scrap this dont need
const { useState, useEffect } = React;

function SentimentScore({ score, label, count, color = "blue" }) {
    const percentage = (score * 100).toFixed(1);
    const sentiment = score > 0 ? 'Bullish' : score < 0 ? 'Bearish' : 'Neutral';
    
    return (
        <div className={`bg-${color}-50 p-6 rounded-lg shadow-sm`}>
            <h3 className={`text-lg font-semibold text-${color}-800`}>{label}</h3>
            <div className="mt-2">
                <div className="text-3xl font-bold">{percentage}%</div>
                {count !== undefined && (
                    <div className="text-sm text-gray-600 mt-1">
                        Based on {count} {label.includes('Twitter') ? 'tweets' : 'posts'}
                    </div>
                )}
                <div className={`mt-2 inline-block px-2 py-1 rounded text-sm ${
                    sentiment === 'Bullish' ? 'bg-green-100 text-green-800' :
                    sentiment === 'Bearish' ? 'bg-red-100 text-red-800' :
                    'bg-gray-100 text-gray-800'
                }`}>
                    {sentiment}
                </div>
            </div>
        </div>
    );
}

function App() {
    const [sentiment, setSentiment] = useState(null);
    const [symbol, setSymbol] = useState('BTC');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);

    const fetchSentiment = async () => {
        try {
            setLoading(true);
            setError(null);
            
            console.log('Fetching sentiment for:', symbol); // Debug log
            
            const response = await fetch(`/api/v1/sentiment/${symbol}`);
            console.log('Response status:', response.status); // Debug log
            
            if (!response.ok) {
                throw new Error(`Failed to fetch sentiment data: ${response.statusText}`);
            }
            
            const data = await response.json();
            console.log('Received data:', data); // Debug log
            
            setSentiment(data);
        } catch (err) {
            console.error('Error fetching sentiment:', err); // Debug log
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchSentiment();
    }, [symbol]);

    return (
        <div className="container mx-auto px-4 py-8">
            <div className="bg-white rounded-lg shadow-lg p-6">
                <div className="flex justify-between items-center mb-6">
                    <div>
                        <h1 className="text-2xl font-bold text-gray-800">
                            Crypto Sentiment Analyzer
                        </h1>
                        <p className="text-gray-600 mt-1">
                            Analyze social media sentiment for cryptocurrencies
                        </p>
                    </div>
                    <div className="flex gap-4">
                        <input
                            type="text"
                            value={symbol}
                            onChange={(e) => setSymbol(e.target.value.toUpperCase())}
                            className="border rounded px-3 py-2 w-32"
                            placeholder="e.g., BTC"
                        />
                        <button
                            onClick={fetchSentiment}
                            disabled={loading}
                            className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600 disabled:opacity-50 flex items-center gap-2"
                        >
                            {loading ? (
                                <>
                                    <div className="animate-spin h-4 w-4 border-2 border-white border-t-transparent rounded-full"></div>
                                    Analyzing...
                                </>
                            ) : (
                                'Analyze'
                            )}
                        </button>
                    </div>
                </div>

                {error && (
                    <div className="bg-red-100 text-red-700 p-4 rounded mb-6">
                        <div className="font-medium">Error</div>
                        {error}
                    </div>
                )}

                {sentiment && !error && (
                    <div className="space-y-6">
                        <SentimentScore 
                            score={sentiment.overall_score}
                            label="Overall Sentiment"
                            count={sentiment.reddit_posts + sentiment.tweets}
                            color="blue"
                        />

                        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                            <SentimentScore 
                                score={sentiment.reddit_score}
                                label="Reddit Sentiment"
                                count={sentiment.reddit_posts}
                                color="green"
                            />
                            <SentimentScore 
                                score={sentiment.twitter_score}
                                label="Twitter Sentiment"
                                count={sentiment.tweets}
                                color="purple"
                            />
                        </div>

                        <div className="text-sm text-gray-500 text-right">
                            Last updated: {new Date(sentiment.timestamp).toLocaleString()}
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}

// Add error boundary
try {
    ReactDOM.render(<App />, document.getElementById('root'));
    console.log('React app rendered successfully');
} catch (error) {
    console.error('Error rendering React app:', error);
    document.getElementById('root').innerHTML = 'Error loading application: ' + error.message;
}