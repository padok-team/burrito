function formatHumanDate(value?: string | number | Date | null): string {
    if (!value) return '';
    const date = value instanceof Date ? value : new Date(value);
    if (isNaN(date.getTime())) return String(value);

    const seconds = Math.floor((Date.now() - date.getTime()) / 1000);
    const rtf = new Intl.RelativeTimeFormat(undefined, { numeric: 'auto' });
    const intervals: [Intl.RelativeTimeFormatUnit, number][] = [
        ['year', 31536000],
        ['month', 2592000],
        ['week', 604800],
        ['day', 86400],
        ['hour', 3600],
        ['minute', 60],
        ['second', 1],
    ];

    for (const [unit, unitSeconds] of intervals) {
        if (Math.abs(seconds) >= unitSeconds || unit === 'second') {
        const valueAmount = Math.round(seconds / unitSeconds);
        return rtf.format(-valueAmount, unit);
        }
    }

    return date.toLocaleString();
}

export { formatHumanDate };
