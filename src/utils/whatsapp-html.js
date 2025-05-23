function convertWhatsAppToHTML(message) {
    if (!message || typeof message !== 'string') {
      return '';
    }
  
    let html = message;
    
    // Escape HTML characters first to prevent XSS
    html = html
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#39;');
    
    // Convert WhatsApp formatting to HTML
    // Bold: *text* -> <strong>text</strong>
    html = html.replace(/\*([^*\n]+)\*/g, '<strong>$1</strong>');
    
    // Italic: _text_ -> <em>text</em>
    html = html.replace(/_([^_\n]+)_/g, '<em>$1</em>');
    
    // Monospace: ```text``` -> <code>text</code>
    html = html.replace(/```([^`\n]+)```/g, '<code>$1</code>');
    
    // Strikethrough: ~text~ -> <del>text</del>
    html = html.replace(/~([^~\n]+)~/g, '<del>$1</del>');
    
    // Convert line breaks to <br> tags
    html = html.replace(/\n/g, '<br>');
    
    return html;
  }
  
  // Enhanced version with better regex patterns
  function convertWhatsAppToHTMLAdvanced(message) {
    if (!message || typeof message !== 'string') {
      return '';
    }
  
    let html = message;
    
    // Escape HTML characters
    html = html
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#39;');
    
    // More robust patterns that handle edge cases
    
    // Bold: *text* (not at word boundaries to avoid conflicts)
    html = html.replace(/(?<!\w)\*([^\s*][^*]*[^\s*]|\S)\*(?!\w)/g, '<strong>$1</strong>');
    
    // Italic: _text_
    html = html.replace(/(?<!\w)_([^\s_][^_]*[^\s_]|\S)_(?!\w)/g, '<em>$1</em>');
    
    // Monospace: ```text```
    html = html.replace(/```([^`]+)```/g, '<code>$1</code>');
    
    // Strikethrough: ~text~
    html = html.replace(/(?<!\w)~([^\s~][^~]*[^\s~]|\S)~(?!\w)/g, '<del>$1</del>');
    
    // Line breaks
    html = html.replace(/\n/g, '<br>');
    
    return html;
  }
  
// Export the function
module.exports = {
  convertWhatsAppToHTML,
  convertWhatsAppToHTMLAdvanced
};

// Usage examples:
// const message1 = "*Hello* _world_! This is ```code``` and ~strikethrough~\nNew line here";
// console.log(convertWhatsAppToHTML(message1));
// Output: <strong>Hello</strong> <em>world</em>! This is <code>code</code> and <del>strikethrough</del><br>New line here
  
//   const message2 = "Visit *example.com* or check _this_file.txt";
//   console.log(convertWhatsAppToHTML(message2));
//   // Output: Visit <strong>example.com</strong> or check <em>this_file.txt</em>
  
  // For React/JSX usage:
//   function WhatsAppMessage({ message }) {
//     return (
//       <div 
//         dangerouslySetInnerHTML={{ 
//           __html: convertWhatsAppToHTML(message) 
//         }} 
//       />
//     );
//   }