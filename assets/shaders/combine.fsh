#version 330

uniform sampler2D tex;
uniform sampler2D tex2;
uniform float power;

in vec2 tex_coord;
out vec4 color;

void main()
{
    vec4 in_color = texture(tex, tex_coord);
    vec4 in_color2 = texture(tex2, tex_coord);
	color = in_color + in_color2 * power;
}