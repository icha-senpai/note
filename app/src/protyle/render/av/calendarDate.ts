export const localDateKey = (date: Date) => `${date.getFullYear().toString().padStart(4, "0")}-${(date.getMonth() + 1).toString().padStart(2, "0")}-${date.getDate().toString().padStart(2, "0")}`;

export const calendarDays = (month: string): string[] => {
    if (!/^\d{4}-(0[1-9]|1[0-2])$/.test(month)) { return []; }
    const [year, index] = month.split("-").map(Number);
    const first = new Date(year, index - 1, 1);
    first.setFullYear(year);
    const offset = (first.getDay() + 6) % 7;
    const last = new Date(year, index, 0);
    last.setFullYear(year);
    const count = Math.ceil((offset + last.getDate()) / 7) * 7;
    first.setDate(1 - offset);
    return Array.from({length: count}, (_, index) => {
        const day = new Date(first);
        day.setDate(first.getDate() + index);
        return localDateKey(day);
    });
};

export const moveCalendarDate = (value: IAVCellDateValue, target: string): IAVCellDateValue => {
    const next = {...value};
    const [year, month, day] = target.split("-").map(Number);
    const start = value?.isNotEmpty ? new Date(value.content) : new Date(year, month - 1, day);
    const old = localDateKey(start);
    start.setFullYear(year, month - 1, day);
    if (!value?.isNotEmpty) { start.setHours(0, 0, 0, 0); }
    next.content = start.getTime();
    next.isNotEmpty = true;
    next.isNotTime = value?.isNotEmpty ? value.isNotTime : true;
    if (value?.isNotEmpty && value.hasEndDate && value.isNotEmpty2) {
        const delta = (Date.parse(target) - Date.parse(old)) / 86400000;
        const end = new Date(value.content2);
        end.setDate(end.getDate() + delta);
        next.content2 = end.getTime();
    }
    return next;
};

export const calendarDateValue = (cell?: IAVCell): IAVCellDateValue =>
    cell?.value?.[cell.valueType as "date"] || {content: 0, isNotEmpty: false, content2: 0, isNotEmpty2: false, hasEndDate: false, isNotTime: true};
