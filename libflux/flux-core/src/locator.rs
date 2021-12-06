use crate::ast::{Position, SourceLocation};

/// Locator makes constructing an SourceLocation from a string simple.
pub struct Locator<'a> {
    source: &'a str,
    lines: Vec<u32>,
}

impl<'a> Locator<'a> {
    /// Create a new Locator for the given source code
    pub fn new(source: &'a str) -> Self {
        let mut lines = Vec::new();
        lines.push(0);
        let ci = source.char_indices();
        for (i, c) in ci {
            match c {
                '\n' => lines.push((i + 1) as u32),
                _ => (),
            }
        }
        Self { source, lines }
    }

    /// Get the SourceLocation for the given start line, start column, end line and end
    /// column.
    #[cfg(test)]
    pub fn get(&self, sl: u32, sc: u32, el: u32, ec: u32) -> SourceLocation {
        SourceLocation {
            file: Some("".to_string()),
            start: Position {
                line: sl,
                column: sc,
            },
            end: Position {
                line: el,
                column: ec,
            },
        }
    }

    pub fn get_src(&self, loc: &SourceLocation) -> Option<&'a str> {
        let SourceLocation {
            start: Position {
                line: sl,
                column: sc,
            },
            end: Position {
                line: el,
                column: ec,
            },
            ..
        } = *loc;
        let start_offset = self.lines.get(sl as usize - 1).expect("line not found") + sc - 1;
        let end_offset = self.lines.get(el as usize - 1).expect("line not found") + ec - 1;
        self.source.get(start_offset as usize..end_offset as usize)
    }
}
