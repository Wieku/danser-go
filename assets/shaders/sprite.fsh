#version 330

uniform sampler2D tex;
uniform vec4 col_tint;

in vec2 tex_coord;
out vec4 color;

void main()
{
    vec4 in_color = texture(tex, tex_coord);
	color = in_color*col_tint;
}