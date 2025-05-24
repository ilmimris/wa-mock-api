const Joi = require('joi');
const { ApiError } = require('./error.middleware');

// Define validation schemas
const messageSchema = Joi.object({
  timestamp: Joi.string().isoDate().required(),
  sender: Joi.string().valid('Bot', 'Customer').required(),
  content: Joi.string().required(),
  recipient_name: Joi.string().optional(),
  recipient_phone: Joi.string().optional()
});

const optionsSchema = Joi.object({
  width: Joi.number().min(300).max(1200).default(400),
  headerDisplay: Joi.string().valid('name', 'phone').default('phone'),
  quality: Joi.string().valid('low', 'medium', 'high').default('high'),
  format: Joi.string().valid('png', 'jpeg', 'webp').default('png')
});

const requestSchema = Joi.object({
  messages: Joi.array().items(messageSchema).min(1).required(),
  options: optionsSchema.optional()
});

/**
 * Validates request body against the schema
 * @param {Object} schema - Joi validation schema
 * @returns {Function} Express middleware function
 */
const validateRequest = (schema) => (req, res, next) => {
  const { error, value } = schema.validate(req.body, { abortEarly: false });
  
  if (error) {
    const errorMessage = error.details.map(detail => detail.message).join(', ');
    return next(new ApiError(400, `Validation error: ${errorMessage}`));
  }
  
  // Replace the request body with the validated value
  req.body = value;
  next();
};

// Export validation middleware for different schemas
module.exports = {
  validateScreenshotRequest: validateRequest(requestSchema),
  messageSchema,
  optionsSchema,
  requestSchema
};
