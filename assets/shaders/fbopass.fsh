#version 330

uniform sampler2D tex;

in vec2 tex_coord;
out vec4 color;

void main()
{
    vec4 in_color = texture(tex, tex_coord);
	color = in_color;
}